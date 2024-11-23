import puppeteer, { Browser } from "puppeteer";
import express, { Request, Response } from "express";


interface MenuItem {
    id: number;
    name: string;
    description: string;
}

interface MenuCategory {
    id: number;
    name: string;
    description: string;
    menu_items: MenuItem[];
}

interface Menu {
    id: number;
    menu_categories: MenuCategory[];
}

interface MenuRespond {
    name: string;
    code: string;
    web_path: string;
    menus: Menu[];
}

const app = express();

// Middleware to parse JSON request bodies
app.use(express.json());

async function scrapeVendorData(
    code: string,
    longitude: number,
    latitude: number
): Promise<MenuRespond | { code: string; error: string }> {
    let browser: Browser | null = null;
    try {
        browser = await puppeteer.launch({
            headless: true,
            args: ["--disable-web-security", "--no-sandbox", "--disable-setuid-sandbox"],
        });

        const page = await browser.newPage();

        await page.setViewport({ width: 1280, height: 800 });
        await page.setUserAgent(
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"
        );

        const url = `https://tw.fd-api.com/api/v5/vendors/${code}?include=menus&longitude=${longitude}&latitude=${latitude}`;
        const response = await page.goto(url, {
            waitUntil: "networkidle0",
            timeout: 30000,
        });

        if (!response || !response.ok()) {
            throw new Error(`Failed to load page: ${response ? response.status() : "No response"}`);
        }

        const pageContent = await page.content();
        const jsonData = extractJSONFromContent(pageContent);

        if (!jsonData || typeof jsonData !== "object") {
            throw new Error(`Malformed JSON response for vendor ${code}`);
        }

        const foodPandaData = jsonData as { data: any };

        if (!foodPandaData.data) {
            return { code, error: "Vendor response missing 'data' field." };
        }

        const menuData = foodPandaData.data;

        // Safeguard `menus` array
        if (!Array.isArray(menuData.menus)) {
            return { code, error: "Invalid or missing 'menus' field in vendor response." };
        }

        // Transform data to `MenuRespond` structure
        const transformedMenus: Menu[] = menuData.menus.map((menu: any) => ({
            id: menu.id,
            menu_categories: Array.isArray(menu.menu_categories) // Safeguard `menu_categories`
                ? menu.menu_categories.map((category: any) => {
                    //   console.log("Raw menu_categories:", category); // Debugging log

                      // Map `product` objects inside each category
                      const transformedItems = Array.isArray(category.products)
                          ? category.products.map((product: any) => ({
                                id: product.id,
                                name: product.name,
                                description: product.description || "",
                                price: product.display_price || product.product_variations?.[0]?.price || 0,
                            }))
                          : []; // Fallback for missing products

                      return {
                          id: category.id,
                          name: category.name,
                          description: category.description,
                          menu_items: transformedItems, // Include transformed items
                      };
                  })
                : [], // Fallback to empty array
        }));

        return {
            name: menuData.chain.name,
            code: menuData.code,
            web_path: menuData.web_path,
            menus: transformedMenus,
        };
    } catch (error) {
        const message = error instanceof Error ? error.message : "Unknown error occurred";
        console.error(`Error scraping vendor ${code}:`, message);
        return { code, error: message };
    } finally {
        if (browser) {
            await browser.close();
        }
    }
}

/**
 * Extracts JSON data from the HTML content of a page.
 * @param htmlContent - The HTML content of the page.
 * @returns The extracted JSON object or null if extraction fails.
 */
function extractJSONFromContent(htmlContent: string): object | null {
    try {
        const jsonStartIndex = htmlContent.indexOf("{");
        const jsonEndIndex = htmlContent.lastIndexOf("}") + 1;

        if (jsonStartIndex === -1 || jsonEndIndex === -1) {
            throw new Error("No JSON found in page content.");
        }

        const jsonString = htmlContent.slice(jsonStartIndex, jsonEndIndex);
        return JSON.parse(jsonString);
    } catch (error) {
        const message = error instanceof Error ? error.message : "Unknown error occurred";
        console.error("Failed to extract JSON:", message);
        return null;
    }
}

app.post("/menu", async (req: Request, res: Response): Promise<void> => {
    const { code: codes, longitude, latitude } = req.body;
    console.log("Received request:", req.body);

    // Validate the request structure
    if (
        !Array.isArray(codes) ||
        codes.length === 0 ||
        typeof longitude !== "number" ||
        typeof latitude !== "number"
    ) {
        res.status(400).send({ error: "Invalid request. Expected { code: [], longitude, latitude } structure." });
        return;
    }

    try {
        // Create scraping promises for each vendor code
        const scrapePromises = codes.map((code) => {
            if (!code) {
                throw new Error("Invalid vendor code.");
            }
            return scrapeVendorData(code, longitude, latitude);
        });

        // Wait for all promises to resolve
        const results = await Promise.all(scrapePromises);

        // Transform results to separate successes and errors
        const transformedResults = results.map((result) => {
            if ("error" in result) {
                return { code: result.code, error: result.error };
            }
            return result;
        });

        res.status(200).send({ success: true, results: transformedResults });
    } catch (error) {
        const message = error instanceof Error ? error.message : "Unknown error occurred";
        console.error("Error during scraping:", message);
        res.status(500).send({ success: false, error: message });
    }
});

app.listen(3001, () => {
    console.log("Server is running on port 3001");
});