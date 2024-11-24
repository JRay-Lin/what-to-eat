import puppeteerExtra from "puppeteer-extra";
import StealthPlugin from "puppeteer-extra-plugin-stealth";
import puppeteer, { HTTPRequest, Browser, Page } from "puppeteer";
import express, { Request, Response } from "express";
import sqlite3 from 'sqlite3';
import { open, Database } from 'sqlite';
import { setTimeout } from "timers/promises";

const cookies = {
  "_pxhd": "50JSRKoL3NzYdJgjDRf7JxUq0oSYutnmhPWkcO4S1j-s8J9OpqGkAmIwNvL1/oh5p5saskcEQGZk5EyTlIL8Vw==:e77-70Sg7SJ3Pd-vQIgmUiNYk5rR33fUDSIDPQOcaGn7DW0Qg5-yx7MvT1XAzj7id5J4btbWERFBoHbz/-KgZTmg98UzjXvjMO16mHpJKACbZs3qNGEs0evbW8aAVKxsgllIhVy2bxyfOkMU8meHew==",
  "_pxvid": "da6e451b-aa86-11ef-ac6d-a4956d1c2277",
  "_px3":"3f37cba6094127a8ff46964fbb40de8ed6d80616fd528c3f4b803549d183a2f8:CSWm5/+WLAHPqWOeqFV5whzKpNp5LPIJYIM6sTPi0Fiu/RCf6P0q7mySbGbtuHwM2Y6LLBqySnZm6Toe2iiDlw==:1000:LH4DeMgxWghyr5EQxTco/bK/Y31UL6jr7DMHF4H2LsBj2TZ52+W8bS0evS4VxKJjeeuiPxozTPIaNDC6B6LbGLeYqnJqAd2WRPQobrNj7Bajbi332zVRVwBYaPN4kC8y3oIsbP25cKMwjIhg30coMX1ErcxhJOLJq4R2OfAm3HuW3PGetTeIAxkPur5ZPo+TpWCuoIMzXNTLomuCqbRNVSGw1UjtePbl/8RDbGuh6iaWfY0vu9OLgNiPZZYvjCuD1/5K1sL4lsQDheedWLPqIw==",
  "__cf_bm": "tEVHzDvMqjkCGsWKfuMKQp1blhB7YhV3najcr.i5haA-1732468475-1.0.1.1-Fw0305Lt5.euJSrt5gUchpzwupOJlge.zx3UbbbLVXR2A5M3f9QoWgzYFIWthEe3VEG1skT25sOveDDl2.72Ow38_w5oxnYSgwV0l3Mpb6Y"
}

// Interfaces
interface MenuRespond {
  name: string;
  code: string;
  web_path: string;
  menus: Menu[];
}

interface Menu {
  id: number;
  menu_categories: MenuCategory[];
}

interface MenuCategory {
  id: number;
  name: string;
  description: string;
  menu_items: MenuItem[];
}

interface MenuItem {
  id: number;
  name: string;
  description: string;
  price: number;
}

// SQLite cache
class MenuCacheDatabase {
    private db: Database | null = null;
    private readonly cacheExpirationHours = 24; // Cache expiration time in hours
  
    async initialize(): Promise<void> {
      this.db = await open({
        filename: 'menu_cache.db',
        driver: sqlite3.Database
      });
  
      await this.createTable();
    }
  
    private async createTable(): Promise<void> {
      if (!this.db) throw new Error('Database not initialized');
  
      await this.db.exec(`
        CREATE TABLE IF NOT EXISTS menu_cache (
          restaurant_code TEXT PRIMARY KEY,
          menu_data TEXT NOT NULL,
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
      `);
    }
  
    async getMenu(code: string): Promise<MenuRespond | null> {
      if (!this.db) throw new Error('Database not initialized');
  
      const query = `
        SELECT menu_data, created_at 
        FROM menu_cache 
        WHERE restaurant_code = ? 
        AND datetime(created_at, '+${this.cacheExpirationHours} hours') > datetime('now')
      `;
  
      const result = await this.db.get(query, [code]);
      
      if (result) {
        return JSON.parse(result.menu_data);
      }
      
      return null;
    }
  
    async saveMenu(code: string, menuData: MenuRespond): Promise<void> {
      if (!this.db) throw new Error('Database not initialized');
  
      const query = `
        INSERT OR REPLACE INTO menu_cache (restaurant_code, menu_data)
        VALUES (?, ?)
      `;
  
      await this.db.run(query, [
        code,
        JSON.stringify(menuData)
      ]);
    }
  
    async close(): Promise<void> {
      if (this.db) {
        await this.db.close();
        this.db = null;
      }
    }
  }
  

// Apply Stealth Plugin
puppeteerExtra.use(StealthPlugin());

class FoodDeliveryAPI {
  private browser: Browser | null = null;
  private menuCache: MenuCacheDatabase;

  constructor() {
    this.menuCache = new MenuCacheDatabase();
  }

  private readonly userAgents = [
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
  ];

  private readonly headers = {
    'accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7',
    'accept-encoding': 'gzip, deflate, br, zstd',
    'accept-language': 'zh-TW,zh;q=0.9,en-US;q=0.8,en;q=0.7',
    'cache-control': 'max-age=0',
    'sec-ch-ua': '"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"',
    'sec-ch-ua-mobile': '?0',
    'sec-ch-ua-platform': '"macOS"',
    'sec-fetch-dest': 'document',
    'sec-fetch-mode': 'navigate',
    'sec-fetch-site': 'none',
    'sec-fetch-user': '?1',
    'upgrade-insecure-requests': '1'
  };

  private readonly cookies = [
    { name: '_pxvid', value: cookies._pxvid },
    { name: '_pxhd', value: cookies._pxhd },
    { name: '__cf_bm', value: cookies.__cf_bm },
    { name: '_px3', value: cookies._px3 }
  ];


  private readonly config = {
    maxRetries: 3,
    retryDelay: 2000,
    requestTimeout: 30000,
    rateLimitDelay: 1500, // Delay between requests
  };

  async initialize(): Promise<void> {
    // Initialize both browser and database
    await Promise.all([
      puppeteerExtra.launch({
        headless: true,
        args: [
          "--disable-web-security",
          "--no-sandbox",
          "--disable-setuid-sandbox",
          "--disable-features=site-per-process",
          "--disable-blink-features=AutomationControlled",
        ],
      }).then(browser => {
        this.browser = browser;
      }),
      this.menuCache.initialize()
    ]);
  }

  async close(): Promise<void> {
    await Promise.all([
      this.browser?.close(),
      this.menuCache.close()
    ]);
    this.browser = null;
  }

  private async createPage(): Promise<Page> {
    if (!this.browser) {
      throw new Error("Browser not initialized");
    }

    const page = await this.browser.newPage();
    
    // Set custom headers
    await page.setExtraHTTPHeaders(this.headers);
    
    // Set user agent
    await page.setUserAgent(this.userAgents[0]);
    
    // Set cookies
    const domain = 'tw.fd-api.com';
    await page.setCookie(...this.cookies.map(cookie => ({
      ...cookie,
      domain,
      path: '/',
      httpOnly: true,
      secure: true
    })));

    // Set up request interception
    await page.setRequestInterception(true);
    page.on("request", (request: HTTPRequest) => {
      if (["image", "stylesheet", "font", "media", "other"].includes(request.resourceType())) {
        request.abort();
      } else {
        // Modify request headers if needed
        const headers = {
          ...request.headers(),
          ...this.headers
        };
        request.continue({ headers });
      }
    });

    return page;
  }

  private async fetchWithRetry(
    page: Page,
    url: string,
    retryCount = 0
  ): Promise<any> {
    try {
      // Add random delay between requests
      await setTimeout(Math.random() * this.config.rateLimitDelay);

      const response = await page.goto(url, {
        waitUntil: "networkidle0",
        timeout: this.config.requestTimeout,
      });

      if (!response) {
        throw new Error("No response received");
      }

      if (response.status() === 429) {
        throw new Error("Rate limit exceeded");
      }

      if (!response.ok()) {
        throw new Error(`HTTP error! status: ${response.status()}`);
      }

      const content = await page.content();
      return this.extractJSONFromContent(content);
    } catch (error) {
      if (retryCount < this.config.maxRetries) {
        console.log(
          `Retrying request (${retryCount + 1}/${this.config.maxRetries})`
        );
        await setTimeout(this.config.retryDelay);
        return this.fetchWithRetry(page, url, retryCount + 1);
      }
      throw error;
    }
  }

  private extractJSONFromContent(htmlContent: string): object {
    try {
      const jsonStartIndex = htmlContent.indexOf("{");
      const jsonEndIndex = htmlContent.lastIndexOf("}") + 1;

      if (jsonStartIndex === -1 || jsonEndIndex === -1) {
        throw new Error("No JSON found in page content.");
      }

      const jsonString = htmlContent.slice(jsonStartIndex, jsonEndIndex);
      return JSON.parse(jsonString);
    } catch (error) {
      throw new Error(
        `JSON extraction failed: ${
          error instanceof Error ? error.message : "Unknown error"
        }`
      );
    }
  }

  async getVendorMenu(
    code: string,
    longitude: number,
    latitude: number
  ): Promise<MenuRespond> {
    // Try to get from cache first
    const cachedMenu = await this.menuCache.getMenu(code);
    if (cachedMenu) {
      console.log(`Retrieved menu for ${code} from cache`);
      return cachedMenu;
    }

    // If not in cache, fetch from API
    const page = await this.createPage();

    try {
      const url = `https://tw.fd-api.com/api/v5/vendors/${code}?include=menus&longitude=${longitude}&latitude=${latitude}`;
      const jsonData = (await this.fetchWithRetry(page, url)) as { data: any };

      if (!jsonData?.data) {
        throw new Error("Invalid API response structure");
      }

      const menuData = jsonData.data;

      const transformedMenu: MenuRespond = {
        name: menuData.chain.name,
        code: menuData.code,
        web_path: menuData.web_path,
        menus: menuData.menus.map((menu: any) => ({
          id: menu.id,
          menu_categories: Array.isArray(menu.menu_categories)
            ? menu.menu_categories.map((category: any) => ({
                id: category.id,
                name: category.name,
                description: category.description || "",
                menu_items: Array.isArray(category.products)
                  ? category.products.map((product: any) => ({
                      id: product.id,
                      name: product.name,
                      description: product.description || "",
                      price:
                        product.display_price ||
                        product.product_variations?.[0]?.price ||
                        0,
                    }))
                  : [],
              }))
            : [],
        })),
      };

      // Save to cache
      await this.menuCache.saveMenu(code, transformedMenu);
      console.log(`Saved menu for ${code} to cache`);

      return transformedMenu;
    } finally {
      await page.close();
    }
  }
}

// Express Setup
const app = express();
app.use(express.json());

const foodDeliveryAPI = new FoodDeliveryAPI();

app.listen(3001, async () => {
  try {
    await foodDeliveryAPI.initialize();
    console.log("Server is running on port 3001");
  } catch (error) {
    console.error("Failed to initialize browser:", error);
    process.exit(1);
  }
});

app.post("/menu", async (req: Request, res: Response): Promise<void> => {
  const { code: codes, longitude, latitude } = req.body;

  if (
    !Array.isArray(codes) ||
    !codes.length ||
    typeof longitude !== "number" ||
    typeof latitude !== "number"
  ) {
    res.status(400).json({ error: "Invalid request parameters" });
    return;
  }

  try {
    const results = await Promise.all(
      codes.map(async (code) => {
        try {
          return await foodDeliveryAPI.getVendorMenu(code, longitude, latitude);
        } catch (error) {
          return {
            code,
            error: error instanceof Error ? error.message : "Unknown error",
          };
        }
      })
    );

    res.status(200).json({ success: true, results });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

process.on("SIGTERM", async () => {
  await foodDeliveryAPI.close();
  process.exit(0);
});
