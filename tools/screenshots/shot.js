// Renders an HTML file from ansi2html.py to a PNG. Usage: node shot.js <in.html> <out.png>
const { chromium } = require('playwright');
(async () => {
  const [src, dst] = process.argv.slice(2);
  const browser = await chromium.launch();
  const page = await browser.newPage({ deviceScaleFactor: 2, viewport: { width: 900, height: 600 } });
  await page.goto('file://' + src);
  await page.waitForTimeout(300);
  await page.locator('.win').screenshot({ path: dst });
  await browser.close();
})();
