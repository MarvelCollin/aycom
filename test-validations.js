// Test script for registration validations
console.log("Starting validation tests...");

const API_URL = "http://localhost:8080/api/users/register";

// Base valid registration data
const validData = {
  name: "Valid User",
  username: "validuser",
  email: "valid@example.com",
  password: "Valid123!",
  confirm_password: "Valid123!",
  gender: "male",
  date_of_birth: "5-15-2000", // Month 5 (June), Day 15, Year 2000
  security_question: "What was the name of your first pet?",
  security_answer: "Fluffy",
  subscribe_to_newsletter: false,
  recaptcha_token: "dev-mode-token"
};

// Test cases
const testCases = [
  {
    name: "Name too short",
    data: { ...validData, name: "Abc" },
    expectedErrors: ["Name must be at least 4 characters"]
  },
  {
    name: "Name with numbers",
    data: { ...validData, name: "User123" },
    expectedErrors: ["Name must not contain symbols or numbers"]
  },
  {
    name: "Invalid email format",
    data: { ...validData, email: "invalid-email" },
    expectedErrors: ["Invalid email format", "invalid email"]
  },
  {
    name: "Short password",
    data: { ...validData, password: "Short1!", confirm_password: "Short1!" },
    expectedErrors: ["Password must be at least 8 characters"]
  },
  {
    name: "Password missing uppercase",
    data: { ...validData, password: "nouppercase123!", confirm_password: "nouppercase123!" },
    expectedErrors: ["Password must contain at least one uppercase letter"]
  },
  {
    name: "Password missing lowercase",
    data: { ...validData, password: "NOLOWER123!", confirm_password: "NOLOWER123!" },
    expectedErrors: ["Password must contain at least one lowercase letter"]
  },
  {
    name: "Password missing number",
    data: { ...validData, password: "NoNumbers!", confirm_password: "NoNumbers!" },
    expectedErrors: ["Password must contain at least one number"]
  },
  {
    name: "Password missing special character",
    data: { ...validData, password: "NoSpecial123", confirm_password: "NoSpecial123" },
    expectedErrors: ["Password must contain at least one special character"]
  },
  {
    name: "Passwords don't match",
    data: { ...validData, password: "Password123!", confirm_password: "DifferentPass123!" },
    expectedErrors: ["Password and confirmation password do not match"]
  },
  {
    name: "Invalid gender",
    data: { ...validData, gender: "other" },
    expectedErrors: ["Gender must be either 'male' or 'female'", "invalid gender"]
  },
  {
    name: "User too young",
    data: { ...validData, date_of_birth: "5-15-2015" }, // 10 years old in 2025
    expectedErrors: ["User must be at least 13 years old"]
  },
  {
    name: "Missing security question",
    data: { ...validData, security_question: "" },
    expectedErrors: ["Security question is required"]
  },
  {
    name: "Missing security answer",
    data: { ...validData, security_answer: "" },
    expectedErrors: ["Security answer is required"]
  },
  {
    name: "Short security answer",
    data: { ...validData, security_answer: "Ab" },
    expectedErrors: ["Security answer must be at least 3 characters"]
  }
];

// Run tests sequentially
async function runTests() {
  console.log("Total test cases:", testCases.length);
  
  for (const [i, test] of testCases.entries()) {
    console.log(`\n[${i + 1}/${testCases.length}] Running test: ${test.name}`);
    
    try {
      const response = await fetch(API_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(test.data)
      });
      
      const result = await response.json();
      console.log("Response status:", response.status);
      console.log("Response body:", JSON.stringify(result, null, 2));
      
      // Check if the response contains expected validation errors
      let passed = false;
      const errorMessage = result.message || (result.error ? result.error.message : '');
      
      if (errorMessage) {
        passed = test.expectedErrors.some(expected => 
          errorMessage.toLowerCase().includes(expected.toLowerCase())
        );
      }
      
      if (passed) {
        console.log("✅ TEST PASSED! Found expected validation error.");
      } else {
        console.log("❌ TEST FAILED! Expected validation errors not found:", test.expectedErrors);
      }
    } catch (error) {
      console.error("Error running test:", error);
    }
    
    // Wait a bit between tests to avoid overwhelming the server
    await new Promise(resolve => setTimeout(resolve, 500));
  }
  
  console.log("\n🏁 All tests completed!");
}

// Start the tests
runTests(); 