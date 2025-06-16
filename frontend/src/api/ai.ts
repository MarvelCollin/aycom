import appConfig from "../config/appConfig";

const AI_SERVICE_URL = appConfig.api.aiServiceUrl || "http://localhost:5000";

export async function predictThreadCategory(content: string) {
  try {

    if (!content || content.trim().length < 10) {
      return {
        category: "general",
        confidence: 0,
        all_categories: {}
      };
    }

    const response = await fetch(`${AI_SERVICE_URL}/predict/category`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ content })
    });

    if (!response.ok) {
      console.warn("Category prediction failed:", response.status, response.statusText);
      return {
        category: "general",
        confidence: 0,
        all_categories: {}
      };
    }

    const data = await response.json();

    return {
      category: data.category || "general",
      confidence: data.confidence || 0,
      all_categories: data.all_categories || {}
    };
  } catch (error) {
    console.error("Error predicting thread category:", error);
    return {
      category: "general",
      confidence: 0,
      all_categories: {}
    };
  }
}

export async function checkAIServiceHealth() {
  try {
    const response = await fetch(`${AI_SERVICE_URL}/health`);

    if (!response.ok) {
      return {
        status: "unhealthy",
        model_loaded: false
      };
    }

    return await response.json();
  } catch (error) {
    console.error("Error checking AI service health:", error);
    return {
      status: "unhealthy",
      model_loaded: false
    };
  }
}