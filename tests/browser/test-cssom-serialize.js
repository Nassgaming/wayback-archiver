const puppeteer = require('puppeteer');

/**
 * Test: CSSOM serialization
 * Verifies that CSS rules injected via CSSStyleSheet.insertRule()
 * are serialized into <style> elements before outerHTML capture.
 */
(async () => {
  console.log('=== CSSOM Serialization Test ===\n');

  let passed = 0;
  let failed = 0;
  const failures = [];

  const browser = await puppeteer.launch({
    headless: true,
    executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const page = await browser.newPage();

  // Create a test page that simulates X/Twitter's CSSOM injection pattern
  await page.setContent(`<!DOCTYPE html>
<html>
<head>
  <style id="static-styles">
    body { margin: 0; }
    .base { display: block; }
  </style>
  <style id="cssom-target"></style>
</head>
<body>
  <div class="r-abc123 r-def456 r-ghi789 base">Test content</div>
</body>
</html>`);

  // Inject CSS rules via CSSOM (simulating what X/Twitter does)
  await page.evaluate(() => {
    const styleEl = document.getElementById('cssom-target');
    const sheet = styleEl.sheet;
    sheet.insertRule('.r-abc123 { color: red; font-size: 16px; }', 0);
    sheet.insertRule('.r-def456 { background-color: blue; padding: 10px; }', 1);
    sheet.insertRule('.r-ghi789 { margin: 20px; border: 1px solid black; }', 2);
    sheet.insertRule('@media (max-width: 768px) { .r-abc123 { font-size: 14px; } }', 3);
  });

  // Test 1: Verify CSSOM rules exist but are NOT in outerHTML before serialization
  const beforeCapture = await page.evaluate(() => {
    const styleEl = document.getElementById('cssom-target');
    const domText = styleEl.textContent;
    const cssomRules = styleEl.sheet.cssRules.length;
    const outerHTML = document.documentElement.outerHTML;
    return {
      domText,
      cssomRules,
      hasR_abc123InHTML: outerHTML.includes('.r-abc123'),
      hasR_def456InHTML: outerHTML.includes('.r-def456'),
    };
  });

  passed++; // test count
  if (beforeCapture.cssomRules === 4 && beforeCapture.domText === '') {
    console.log(`  [PASS] CSSOM has ${beforeCapture.cssomRules} rules, DOM <style> is empty (as expected)`);
  } else {
    failed++;
    const msg = `Expected 4 CSSOM rules and empty DOM, got ${beforeCapture.cssomRules} rules and "${beforeCapture.domText}"`;
    failures.push(msg);
    console.log(`  [FAIL] ${msg}`);
  }

  passed++;
  if (!beforeCapture.hasR_abc123InHTML) {
    console.log('  [PASS] outerHTML does NOT contain .r-abc123 before serialization');
  } else {
    failed++;
    failures.push('outerHTML contains .r-abc123 before serialization');
    console.log('  [FAIL] outerHTML already contains .r-abc123 before serialization');
  }

  // Now inject and run the serializeCSSOM function
  await page.evaluate(() => {
    function serializeCSSOM() {
      let serialized = 0;
      try {
        for (let i = 0; i < document.styleSheets.length; i++) {
          const sheet = document.styleSheets[i];
          if (sheet.href) continue;

          let rules;
          try {
            rules = sheet.cssRules;
          } catch {
            continue;
          }

          const ownerNode = sheet.ownerNode;
          if (!(ownerNode instanceof HTMLStyleElement)) continue;

          const parts = [];
          for (let j = 0; j < rules.length; j++) {
            parts.push(rules[j].cssText);
          }
          const cssom = parts.join('\n');

          const domText = ownerNode.textContent || '';
          if (cssom.length > domText.length) {
            ownerNode.textContent = cssom;
            serialized++;
          }
        }
      } catch (e) {
        console.warn('CSSOM serialization error:', e);
      }
      return serialized;
    }
    window.__serializedCount = serializeCSSOM();
  });

  // Test 2: Verify serialization happened
  const serializedCount = await page.evaluate(() => window.__serializedCount);
  passed++;
  if (serializedCount === 1) {
    console.log(`  [PASS] Serialized ${serializedCount} stylesheet (the CSSOM-injected one)`);
  } else {
    failed++;
    const msg = `Expected 1 serialized stylesheet, got ${serializedCount}`;
    failures.push(msg);
    console.log(`  [FAIL] ${msg}`);
  }

  // Test 3: Verify outerHTML now contains the CSSOM rules
  const afterCapture = await page.evaluate(() => {
    const outerHTML = document.documentElement.outerHTML;
    return {
      hasR_abc123: outerHTML.includes('.r-abc123'),
      hasR_def456: outerHTML.includes('.r-def456'),
      hasR_ghi789: outerHTML.includes('.r-ghi789'),
      hasMediaQuery: outerHTML.includes('max-width'),
      hasColorRed: outerHTML.includes('color: red') || outerHTML.includes('color:red'),
      hasBgBlue: outerHTML.includes('background-color: blue') || outerHTML.includes('background-color:blue'),
    };
  });

  const checks = [
    ['r-abc123 class rule', afterCapture.hasR_abc123],
    ['r-def456 class rule', afterCapture.hasR_def456],
    ['r-ghi789 class rule', afterCapture.hasR_ghi789],
    ['@media query', afterCapture.hasMediaQuery],
    ['color: red property', afterCapture.hasColorRed],
    ['background-color: blue property', afterCapture.hasBgBlue],
  ];

  for (const [name, ok] of checks) {
    passed++;
    if (ok) {
      console.log(`  [PASS] outerHTML contains ${name} after serialization`);
    } else {
      failed++;
      failures.push(`outerHTML missing ${name} after serialization`);
      console.log(`  [FAIL] outerHTML missing ${name} after serialization`);
    }
  }

  // Test 4: Static styles should NOT be overwritten (they already have content)
  const staticStyleCheck = await page.evaluate(() => {
    const el = document.getElementById('static-styles');
    const text = el.textContent;
    return {
      hasBody: text.includes('body'),
      hasBase: text.includes('.base'),
      // Should NOT have been overwritten with CSSOM version
      length: text.length,
    };
  });

  passed++;
  if (staticStyleCheck.hasBody && staticStyleCheck.hasBase) {
    console.log('  [PASS] Static <style> content preserved (not overwritten)');
  } else {
    failed++;
    failures.push('Static <style> content was overwritten');
    console.log('  [FAIL] Static <style> content was overwritten');
  }

  // Test 5: Simulate full capture flow (freeze + outerHTML)
  const page2 = await browser.newPage();
  await page2.setContent(`<!DOCTYPE html>
<html>
<head>
  <style id="injected"></style>
</head>
<body>
  <div class="r-test1 r-test2">Content</div>
</body>
</html>`);

  // Inject rules then capture like the real archiver does
  const capturedHTML = await page2.evaluate(() => {
    // Simulate site injecting CSS via CSSOM
    const sheet = document.getElementById('injected').sheet;
    sheet.insertRule('.r-test1 { display: flex; align-items: center; }', 0);
    sheet.insertRule('.r-test2 { position: absolute; top: 0; left: 0; }', 1);

    // Simulate serializeCSSOM
    for (let i = 0; i < document.styleSheets.length; i++) {
      const s = document.styleSheets[i];
      if (s.href) continue;
      let rules;
      try { rules = s.cssRules; } catch { continue; }
      const ownerNode = s.ownerNode;
      if (!(ownerNode instanceof HTMLStyleElement)) continue;
      const parts = [];
      for (let j = 0; j < rules.length; j++) parts.push(rules[j].cssText);
      const cssom = parts.join('\n');
      const domText = ownerNode.textContent || '';
      if (cssom.length > domText.length) ownerNode.textContent = cssom;
    }

    // Capture like the real archiver
    return document.documentElement.outerHTML;
  });

  passed++;
  const hasAllRules = capturedHTML.includes('.r-test1') &&
                      capturedHTML.includes('.r-test2') &&
                      capturedHTML.includes('display: flex') &&
                      capturedHTML.includes('position: absolute');
  if (hasAllRules) {
    console.log('  [PASS] Full capture flow: CSSOM rules present in captured HTML');
  } else {
    failed++;
    failures.push('Full capture flow: CSSOM rules missing from captured HTML');
    console.log('  [FAIL] Full capture flow: CSSOM rules missing from captured HTML');
  }

  await browser.close();

  // Summary
  const total = passed;
  const actualPassed = passed - failed;
  console.log('\n========================================');
  console.log('         TEST RESULTS SUMMARY');
  console.log('========================================');
  console.log(`Total:  ${total}`);
  console.log(`Passed: ${actualPassed}`);
  console.log(`Failed: ${failed}`);

  if (failures.length > 0) {
    console.log('\n--- Failures ---');
    failures.forEach((f, i) => console.log(`${i + 1}. ${f}`));
  }

  console.log('\n========================================');
  process.exit(failed > 0 ? 1 : 0);
})();
