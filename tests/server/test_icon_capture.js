const fs = require('fs');
const path = require('path');

async function testIconCapture() {
  console.log('Testing icon resource capture...\n');

  // 读取测试HTML
  const htmlPath = path.join(__dirname, 'test_icons.html');
  const html = fs.readFileSync(htmlPath, 'utf-8');

  // 模拟浏览器扩展发送的数据
  const captureData = {
    url: 'https://example.com/',
    title: 'Test Page - Icon Resources',
    html: html,
    timestamp: Date.now(),
    resources: [
      {
        url: 'https://example.com/icon.svg',
        type: 'image',
        content: Buffer.from('<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"><circle cx="50" cy="50" r="40" fill="blue"/></svg>').toString('base64')
      },
      {
        url: 'https://example.com/icons/zhihu.png',
        type: 'image',
        content: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8DwHwAFBQIAX8jx0gAAAABJRU5ErkJggg=='
      },
      {
        url: 'https://example.com/icons/weibo.png',
        type: 'image',
        content: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+P+/HgAE/gL+kR/GGQAAAABJRU5ErkJggg=='
      },
      {
        url: 'https://example.com/icons/coolapk.png',
        type: 'image',
        content: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=='
      },
      {
        url: 'https://example.com/icons/wallstreetcn.png',
        type: 'image',
        content: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYGD4DwABBAEAW9JJRQAAAABJRU5ErkJggg=='
      }
    ]
  };

  try {
    // 发送到服务器
    const response = await fetch('http://localhost:8080/api/archive', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(captureData)
    });

    const result = await response.json();
    console.log('Capture result:', result);

    if (result.status === 'success') {
      console.log('\n✓ Page captured successfully!');

      // 获取最新的页面ID
      const pagesResponse = await fetch('http://localhost:8080/api/pages');
      const pages = await pagesResponse.json();
      const latestPage = pages[0];

      console.log(`\nLatest page ID: ${latestPage.id}`);
      console.log(`Page URL: ${latestPage.url}`);
      console.log(`\nYou can view it at: http://localhost:8080/view/${latestPage.id}`);

      return latestPage.id;
    } else {
      console.error('\n✗ Capture failed');
      return null;
    }
  } catch (error) {
    console.error('Error:', error.message);
    return null;
  }
}

testIconCapture();
