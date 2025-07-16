// Simple test script to verify the Go service is working
const http = require('http');

// Test data
const testEmail = {
  to: "test@example.com",
  subject: "Test Email",
  body: "This is a test email from the queue service."
};

// Function to make HTTP requests
function makeRequest(options, data) {
  return new Promise((resolve, reject) => {
    const req = http.request(options, (res) => {
      let body = '';
      res.on('data', (chunk) => body += chunk);
      res.on('end', () => {
        resolve({
          statusCode: res.statusCode,
          headers: res.headers,
          body: body
        });
      });
    });
    
    req.on('error', reject);
    
    if (data) {
      req.write(JSON.stringify(data));
    }
    req.end();
  });
}

// Test functions
async function testHealthCheck() {
  console.log('\n🔍 Testing health check...');
  try {
    const response = await makeRequest({
      hostname: 'localhost',
      port: 8080,
      path: '/health',
      method: 'GET'
    });
    
    console.log(`✅ Health check: ${response.statusCode} - ${response.body}`);
    return response.statusCode === 200;
  } catch (error) {
    console.log(`❌ Health check failed: ${error.message}`);
    return false;
  }
}

async function testSendEmail() {
  console.log('\n📧 Testing send email...');
  try {
    const response = await makeRequest({
      hostname: 'localhost',
      port: 8080,
      path: '/api/v1/send-email',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      }
    }, testEmail);
    
    console.log(`✅ Send email: ${response.statusCode}`);
    console.log(`Response: ${response.body}`);
    return response.statusCode === 202;
  } catch (error) {
    console.log(`❌ Send email failed: ${error.message}`);
    return false;
  }
}

async function testStats() {
  console.log('\n📊 Testing stats...');
  try {
    const response = await makeRequest({
      hostname: 'localhost',
      port: 8080,
      path: '/api/v1/stats',
      method: 'GET'
    });
    
    console.log(`✅ Stats: ${response.statusCode}`);
    console.log(`Response: ${response.body}`);
    return response.statusCode === 200;
  } catch (error) {
    console.log(`❌ Stats failed: ${error.message}`);
    return false;
  }
}

async function testInvalidEmail() {
  console.log('\n❌ Testing invalid email...');
  try {
    const invalidEmail = {
      to: "invalid-email",
      subject: "Test",
      body: "Test"
    };
    
    const response = await makeRequest({
      hostname: 'localhost',
      port: 8080,
      path: '/api/v1/send-email',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      }
    }, invalidEmail);
    
    console.log(`✅ Invalid email validation: ${response.statusCode}`);
    console.log(`Response: ${response.body}`);
    return response.statusCode === 422;
  } catch (error) {
    console.log(`❌ Invalid email test failed: ${error.message}`);
    return false;
  }
}

// Main test runner
async function runTests() {
  console.log('🚀 Starting Go Email Queue Service Tests...');
  console.log('Make sure the Go service is running on port 8080');
  
  // Wait a moment for service to be ready
  await new Promise(resolve => setTimeout(resolve, 2000));
  
  const tests = [
    testHealthCheck,
    testSendEmail,
    testStats,
    testInvalidEmail
  ];
  
  let passed = 0;
  let total = tests.length;
  
  for (const test of tests) {
    if (await test()) {
      passed++;
    }
    await new Promise(resolve => setTimeout(resolve, 1000)); // Wait between tests
  }
  
  console.log(`\n📋 Test Results: ${passed}/${total} tests passed`);
  
  if (passed === total) {
    console.log('🎉 All tests passed! The service is working correctly.');
  } else {
    console.log('⚠️  Some tests failed. Check the service logs.');
  }
}

// Run tests
runTests().catch(console.error);