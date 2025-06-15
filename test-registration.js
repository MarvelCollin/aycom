// Test script for registration endpoint validation
const fetch = require('node-fetch');

// Update API URL to use the Docker container's exposed port
// Using correct endpoint format from appConfig.ts
const API_URL = "http://localhost:8083/api/v1/users/register";

// Test cases with different validation errors
const testCases = [
  {
    name: "Empty form",
    data: {
      name: "",
      username: "",
      email: "",
      password: "",
      confirm_password: "",
      gender: "",
      date_of_birth: "",
      security_question: "",
      security_answer: "",
      recaptcha_token: "dev-mode-token"
    },
    expectedErrors: ["required", "validation"]
  },
  {
    name: "Short name",
    data: {
      name: "Abc",
      username: "validuser",
      email: "valid@example.com",
      password: "Password123!",
      confirm_password: "Password123!",
      gender: "male",
      date_of_birth: "5-15-2000",
      security_question: "What was your childhood nickname?",
      security_answer: "Nickname",
      recaptcha_token: "dev-mode-token"
    },
    expectedErrors: ["name", "must be at least 4 characters"]
  },
  {
    name: "Name with numbers",
    data: {
      name: "User123",
      username: "validuser",
      email: "valid@example.com",
      password: "Password123!",
      confirm_password: "Password123!",
      gender: "male",
      date_of_birth: "5-15-2000",
      security_question: "What was your childhood nickname?",
      security_answer: "Nickname",
      recaptcha_token: "dev-mode-token"
    },
    expectedErrors: ["name", "must not contain"]
  },
  {
    name: "Invalid email",
    data: {
      name: "Valid User",
      username: "validuser",
      email: "invalid-email",
      password: "Password123!",
      confirm_password: "Password123!",
      gender: "male",
      date_of_birth: "5-15-2000",
      security_question: "What was your childhood nickname?",
      security_answer: "Nickname",
      recaptcha_token: "dev-mode-token"
    },
    expectedErrors: ["email", "invalid"]
  },
  {
    name: "Weak password (missing upper)",
    data: {
      name: "Valid User",
      username: "validuser",
      email: "valid@example.com",
      password: "password123!",
      confirm_password: "password123!",
      gender: "male",
      date_of_birth: "5-15-2000",
      security_question: "What was your childhood nickname?",
      security_answer: "Nickname",
      recaptcha_token: "dev-mode-token"
    },
    expectedErrors: ["uppercase"]
  },
  {
    name: "Password mismatch",
    data: {
      name: "Valid User",
      username: "validuser",
      email: "valid@example.com",
      password: "Password123!",
      confirm_password: "DifferentPass123!",
      gender: "male",
      date_of_birth: "5-15-2000",
      security_question: "What was your childhood nickname?",
      security_answer: "Nickname",
      recaptcha_token: "dev-mode-token"
    },
    expectedErrors: ["match", "password"]
  },
  {
    name: "Invalid gender",
    data: {
      name: "Valid User",
      username: "validuser",
      email: "valid@example.com",
      password: "Password123!",
      confirm_password: "Password123!",
      gender: "other",
      date_of_birth: "5-15-2000",
      security_question: "What was your childhood nickname?",
      security_answer: "Nickname",
      recaptcha_token: "dev-mode-token"
    },
    expectedErrors: ["gender", "male", "female"]
  },
  {
    name: "User too young",
    data: {
      name: "Valid User",
      username: "validuser",
      email: "valid@example.com",
      password: "Password123!",
      confirm_password: "Password123!",
      gender: "male",
      date_of_birth: "5-15-2020",
      security_question: "What was your childhood nickname?",
      security_answer: "Nickname",
      recaptcha_token: "dev-mode-token"
    },
    expectedErrors: ["13 years", "age"]
  },
  {
    name: "Short security answer",
    data: {
      name: "Valid User",
      username: "validuser",
      email: "valid@example.com",
      password: "Password123!",
      confirm_password: "Password123!",
      gender: "male",
      date_of_birth: "5-15-2000",
      security_question: "What was your childhood nickname?",
      security_answer: "Ni",
      recaptcha_token: "dev-mode-token"
    },
    expectedErrors: ["security answer", "at least 3"]
  },
  // Finally, test a fully valid registration to make sure it works
  {
    name: "Valid registration",
    data: {
      name: "Test User",
      username: "testuser" + Math.floor(Math.random() * 10000), // Random suffix to avoid duplicates
      email: `test${Math.floor(Math.random() * 10000)}@example.com`, // Random email to avoid duplicates
      password: "Password123!",
      confirm_password: "Password123!",
      gender: "male",
      date_of_birth: "5-15-2000",
      security_question: "What was your childhood nickname?",
      security_answer: "Nickname",
      recaptcha_token: "dev-mode-token"
    },
    expectedErrors: null // Should pass validation
  }
];

// Run the tests
async function runTests() {
  console.log("Starting registration validation tests...\n");
  
  for (let i = 0; i < testCases.length; i++) {
    const test = testCases[i];
    console.log(`Test ${i+1}/${testCases.length}: ${test.name}`);
    
    try {
      const response = await fetch(API_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(test.data),
      });
      
      const responseData = await response.json();
      const statusMessage = `Status: ${response.status} ${response.ok ? '(OK)' : '(Error)'}`;
      console.log(statusMessage);
      
      // Check if we're expecting this to be valid or invalid
      if (test.expectedErrors === null) {
        // This should be a successful registration
        if (response.ok && responseData.success) {
          console.log("✅ SUCCESS: Valid registration was accepted\n");
        } else {
          console.log("❌ FAIL: Valid registration was rejected");
          console.log(JSON.stringify(responseData, null, 2) + "\n");
        }
      } else {
        // This should be a validation error
        if (!response.ok || !responseData.success) {
          // Check if expected error messages are in the response
          const errorMessage = responseData.message || 
                              (responseData.error ? responseData.error.message : "");
          
          const hasExpectedErrors = test.expectedErrors.every(errorText => 
            errorMessage.toLowerCase().includes(errorText.toLowerCase())
          );
          
          if (hasExpectedErrors) {
            console.log("✅ SUCCESS: Validation error detected correctly");
            console.log(`Error message: ${errorMessage}\n`);
          } else {
            console.log("⚠️ WARNING: Response has error but not the expected validation messages");
            console.log(`Expected to find: ${test.expectedErrors.join(", ")}`);
            console.log(`Actual message: ${errorMessage}\n`);
          }
        } else {
          console.log("❌ FAIL: Invalid input was incorrectly accepted");
          console.log(JSON.stringify(responseData, null, 2) + "\n");
        }
      }
    } catch (error) {
      console.error(`❌ ERROR running test "${test.name}":`, error.message);
    }
  }
  
  console.log("All tests completed!");
}

// Run the tests
runTests().catch(e => console.error("Test error:", e)); 