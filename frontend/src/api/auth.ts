import appConfig from "../config/appConfig";
import { getAuthToken } from "../utils/auth";
import { createLoggerWithPrefix } from "../utils/logger";

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix("AuthAPI");

async function handleApiResponse(response: Response, errorMessage: string = "Operation failed") {
  
  if (!response.ok) {
    logger.error(`Error status: ${response.status} ${response.statusText}`);

    try {
      const errorData = await response.json();
      logger.error("Error data:", errorData);

      
      let validationErrors = errorData.validation_errors || {};

      
      if (errorData.error && errorData.error.fields && Object.keys(errorData.error.fields).length > 0) {
        validationErrors = { ...validationErrors, ...errorData.error.fields };
      }

      
      const errorDetails = errorData.message ||
                        (errorData.error ? errorData.error.message : null) ||
                        errorMessage;

      const errorCode = errorData.code ||
                     (errorData.error ? errorData.error.code : null) ||
                     "ERROR";

      
      if (typeof errorDetails === "string" &&
          Object.keys(validationErrors).length === 0 &&
          (errorDetails.includes("Validation failed:") || errorDetails.includes("Key:"))) {

        
        const validationMessages = errorDetails
          .replace("Validation failed: ", "")
          .split(";")
          .map(msg => msg.trim())
          .filter(msg => msg);

        
        validationMessages.forEach(msg => {
          if (msg.toLowerCase().includes("name ")) validationErrors["name"] = msg;
          else if (msg.toLowerCase().includes("username ")) validationErrors["username"] = msg;
          else if (msg.toLowerCase().includes("email ")) validationErrors["email"] = msg;
          else if (msg.toLowerCase().includes("password ") && !msg.toLowerCase().includes("confirmation"))
            validationErrors["password"] = msg;
          else if (msg.toLowerCase().includes("password") && msg.toLowerCase().includes("match"))
            validationErrors["confirm_password"] = msg;
          else if (msg.toLowerCase().includes("gender ")) validationErrors["gender"] = msg;
          else if (msg.toLowerCase().includes("date of birth") || msg.toLowerCase().includes("13 year"))
            validationErrors["date_of_birth"] = msg;
          else if (msg.toLowerCase().includes("security question")) validationErrors["security_question"] = msg;
          else if (msg.toLowerCase().includes("security answer")) validationErrors["security_answer"] = msg;
        });
      }

      
      return {
        success: false, 
        error: {
          code: errorCode,
          message: errorDetails
        },
        validation_errors: Object.keys(validationErrors).length > 0 ? validationErrors : undefined
      };
    } catch (parseError) {
      logger.error("Failed to parse error response:", parseError);
      
      return {
        success: false,
        error: {
          code: "PARSE_ERROR",
          message: `${errorMessage} with status: ${response.status}`
        }
      };
    }
  }

  try {
    const data = await response.json();
    logger.debug("Successful response with keys:", Object.keys(data));

    
    if (data.success === undefined) {
      data.success = true;
    }

    return data;
  } catch (parseError) {
    logger.error("Failed to parse successful response:", parseError);
    
    return {
      success: true,
      message: "Operation completed successfully"
    };
  }
}

function standardizeUserResponse(data: any) {
  const standardized = {
    user: data.user ? {
      id: data.user.id,
      username: data.user.username,
      name: data.user.name || data.user.display_name,
      email: data.user.email,
      profile_picture_url: data.user.profile_picture_url || data.user.profilePictureUrl,
      is_verified: data.user.is_verified || data.user.verified || false,
      is_admin: data.user.is_admin || data.user.admin || false
    } : data.user,
    access_token: data.access_token || data.accessToken || data.token,
    refresh_token: data.refresh_token || data.refreshToken,
    expires_in: data.expires_in || data.expiresIn,
    token_type: data.token_type || data.tokenType || "bearer",
    user_id: data.user_id || data.userId || (data.user ? data.user.id : null)
  };

  
  if (standardized.access_token) {
    const tokenPreview = standardized.access_token.substring(0, 10) + "..." +
      standardized.access_token.substring(standardized.access_token.length - 5);
    logger.info(`Standardized response has token: ${tokenPreview}`);
  } else {
    logger.warn("Standardized response has no token");
  }

  return standardized;
}

export async function login(email: string, password: string, recaptchaToken: string | null = null) {
  logger.info(`Attempting to login with email: ${email.substring(0, 3)}...${email.split("@")[1]}`);

  try {
    const response = await fetch(`${API_BASE_URL}/users/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email,
        password,
        recaptcha_token: recaptchaToken || (import.meta.env.DEV ? "dev-mode-token" : "")
      }),
      credentials: "include",
    });

    logger.info(`Login response status: ${response.status} ${response.statusText}`);
    const data = await handleApiResponse(response, "Login failed");
    logger.info(`Login successful, response has fields: ${Object.keys(data).join(", ")}`);

    if (data.access_token) {
      const tokenPreview = data.access_token.substring(0, 10) + "..." +
        data.access_token.substring(data.access_token.length - 5);
      logger.info(`Raw login response has token: ${tokenPreview}, user_id: ${data.user_id}`);
    }

    return standardizeUserResponse(data);
  } catch (error) {
    logger.error("Login exception:", error);
    throw error;
  }
}

export async function register(data: Record<string, any>) {
  logger.info("Sending registration request...");

  try {
    const response = await fetch(`${API_BASE_URL}/users/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
      credentials: "include",
    });

    logger.info(`Registration response status: ${response.status} ${response.statusText}`);

    const result = await handleApiResponse(response, "Registration failed");

    
    if (!response.ok && result.success === true) {
      logger.warn("Correcting inconsistent success flag: API reported success but HTTP status indicates failure");
      result.success = false;
    }

    return result;
  } catch (error) {
    logger.error("Registration request error:", error);
    throw error;
  }
}

export async function refreshToken(refreshToken: string) {
  const token = getAuthToken();

  const response = await fetch(`${API_BASE_URL}/auth/refresh-token`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": token ? `Bearer ${token}` : ""
    },
    body: JSON.stringify({ refresh_token: refreshToken }),
    credentials: "include",
  });

  return handleApiResponse(response, "Token refresh failed");
}

export async function verifyEmail(email: string, verificationCode: string) {
  const response = await fetch(`${API_BASE_URL}/auth/verify-email`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, verification_code: verificationCode }),
    credentials: "include",
  });

  return handleApiResponse(response, "Email verification failed");
}

export async function resendVerification(email: string) {
  const response = await fetch(`${API_BASE_URL}/auth/resend-verification`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
    credentials: "include",
  });

  return handleApiResponse(response, "Resend verification failed");
}

export async function googleLogin(tokenId: string) {
  try {
    const logger = createLoggerWithPrefix("GoogleLogin");
    logger.info(`Sending Google token to backend API for verification (token length: ${tokenId.length})`);

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000);

    const response = await fetch(`${API_BASE_URL}/auth/google`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token_id: tokenId }),
      credentials: "include",
      signal: controller.signal
    });

    clearTimeout(timeoutId);

    logger.info(`Google API response status: ${response.status} ${response.statusText}`);

    if (!response.ok) {
      
      try {
        const errorData = await response.json();
        logger.error("Google login error response:", errorData);
        throw new Error(errorData.message || `Google login failed with status: ${response.status}`);
      } catch (parseError) {
        logger.error("Failed to parse error response:", parseError);
        throw new Error(`Google login failed with status: ${response.status}`);
      }
    }

    const data = await handleApiResponse(response, "Google login failed");
    logger.info("Google login raw response keys:", Object.keys(data));

    
    const standardizedData = {
      ...standardizeUserResponse(data),
      success: response.ok,
      is_new_user: data.is_new_user || false,
      message: data.message || (response.ok ? "Google login successful" : "Google login failed")
    };

    logger.info("Google login processed with standardized response");
    return standardizedData;
  } catch (error) {
    console.error("Google login request error:", error);
    if (error instanceof Error && error.name === "AbortError") {
      throw new Error("Request timed out. The server might be down or not responding.");
    }
    throw error;
  }
}

export async function createAdminUser(data: Record<string, any>) {
  try {
    console.log("Creating admin user with data:", data);

    const response = await fetch(`${API_BASE_URL}/users/register`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        ...data,
        is_admin: true,
        is_verified: true
      }),
      credentials: "include",
    });

    return handleApiResponse(response, "Admin user creation failed");
  } catch (error) {
    console.error("Admin user creation failed:", error);
    throw error;
  }
}