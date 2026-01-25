import { test, expect } from '@playwright/test';

test.describe('gowasm-bindgen Example', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    // Wait for WASM to be ready
    await expect(page.locator('#status')).toHaveClass(/ready/);
  });

  test('greet returns greeting message', async ({ page }) => {
    await page.fill('#greetName', 'Playwright');
    await page.click('#greetBtn');
    await expect(page.locator('#greetResult')).toContainText('Hello, Playwright!');
  });

  test('calculate performs arithmetic', async ({ page }) => {
    await page.fill('#calcA', '10');
    await page.fill('#calcB', '5');
    await page.selectOption('#calcOp', 'add');
    await page.click('#calcBtn');
    await expect(page.locator('#calcResult')).toContainText('15');
  });

  test('formatUser returns user object', async ({ page }) => {
    await page.fill('#userName', 'Alice');
    await page.fill('#userAge', '30');
    await page.check('#userActive');
    await page.click('#formatBtn');
    const result = page.locator('#formatResult');
    await expect(result).toContainText('displayName');
    await expect(result).toContainText('Alice');
  });

  test('sumNumbers parses and sums', async ({ page }) => {
    await page.fill('#numbersInput', '1, 2, 3, 4, 5');
    await page.click('#sumBtn');
    await expect(page.locator('#sumResult')).toContainText('15');
  });

  test('validateEmail returns validation result', async ({ page }) => {
    await page.fill('#emailInput', 'user@example.com');
    await page.click('#emailBtn');
    await expect(page.locator('#emailResult')).toContainText('"valid": true');
  });

  test('divide performs division', async ({ page }) => {
    await page.fill('#divideA', '10');
    await page.fill('#divideB', '2');
    await page.click('#divideBtn');
    await expect(page.locator('#divideResult')).toContainText('10 / 2 = 5');
  });

  test('divide throws on division by zero', async ({ page }) => {
    await page.fill('#divideA', '10');
    await page.fill('#divideB', '0');
    await page.click('#divideBtn');
    await expect(page.locator('#divideResult')).toContainText('Error');
  });

  test('hashData computes hash', async ({ page }) => {
    await page.fill('#hashInput', 'Hello, WASM!');
    await page.click('#hashBtn');
    await expect(page.locator('#hashResult')).toContainText('Hash: 0x');
  });

  test('processNumbers doubles numbers', async ({ page }) => {
    await page.fill('#processInput', '1, 2, 3');
    await page.click('#processBtn');
    await expect(page.locator('#processResult')).toContainText('[2, 4, 6]');
  });

  test('forEach invokes callback for each item', async ({ page }) => {
    await page.fill('#forEachInput', 'apple, banana, cherry');
    await page.click('#forEachBtn');
    const result = page.locator('#forEachResult');
    await expect(result).toContainText('[0] apple');
    await expect(result).toContainText('[1] banana');
    await expect(result).toContainText('[2] cherry');
  });
});
