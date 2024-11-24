import axios from 'axios';

// Replace these variables with your dynamic values
const latitude = 24.1781713; 
const longitude = 120.6494391; 
const cuisine = ""; 

// Build the URL with dynamic parameters
const url = `https://disco.deliveryhero.io/listing/api/v1/pandora/vendors?latitude=${latitude}&longitude=${longitude}&language_id=6&include=characteristics&dynamic_pricing=0&configuration=Original&country=TW&customer_id=&customer_hash=&budgets=&cuisine=${cuisine}&sort=&food_characteristic=&use_free_delivery_label=false&vertical=restaurants&limit=99999&offset=0&customer_type=regular`;

const headers = {
    'x-disco-client-id': 'web', // Add the custom header
  };

// Function to make the GET request
async function fetchVendors() {
    try {
        const response = await axios.get(url, { headers });
    
        // map codes into a list
        const codes:[string] = response.data.data.items.map((item: any) => item.code);
        const codesString = codes.map((code: string) => `"${code}"`).join(', ');
        // console.log('Codes:', codesString);

        // map every 10 codeString into a list
        const bigList: string[] = [];
        for (let i = 0; i < codes.length; i += 10) {
            const chunk = codes.slice(i, i + 10);
            const chunkString = chunk.map((code: string) => `"${code}"`).join(', ');
            bigList.push(chunkString);
        }

        console.log('Codes:', bigList)

        // Save bigList into file
        const fs = require('fs');
        fs.writeFileSync('codes.txt', bigList.join('\n'));

        // Cuisines list
        console.log('Cuisines:', response.data.data.aggregations.cuisines);
      } catch (error:any) {
        console.error('Error fetching vendors:', error.message);
      }
    }


// Execute the function
fetchVendors();