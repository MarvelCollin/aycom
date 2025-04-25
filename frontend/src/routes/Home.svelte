<script lang="ts">
  // Home page component
</script>

<div class="home-page">
  <section class="hero">
    <div class="container">
      <div class="hero-content">
        <h1>Welcome to AYCOM</h1>
        <p>Your trusted online marketplace for all your needs</p>
        <div class="cta-buttons">
          <a href="/products" class="primary-button">Browse Products</a>
          <a href="/register" class="secondary-button">Join Now</a>
        </div>
      </div>
    </div>
  </section>
  
  <section class="features">
    <div class="container">
      <h2 class="section-title">Why Choose AYCOM?</h2>
      <div class="features-grid">
        <div class="feature-card">
          <div class="feature-icon">🚚</div>
          <h3>Fast Delivery</h3>
          <p>Get your products delivered to your doorstep in no time.</p>
        </div>
        
        <div class="feature-card">
          <div class="feature-icon">💎</div>
          <h3>Quality Products</h3>
          <p>All our products are carefully selected for quality.</p>
        </div>
        
        <div class="feature-card">
          <div class="feature-icon">🔒</div>
          <h3>Secure Payments</h3>
          <p>Shop with confidence with our secure payment options.</p>
        </div>
        
        <div class="feature-card">
          <div class="feature-icon">🌍</div>
          <h3>Global Reach</h3>
          <p>Connect with sellers and buyers from around the world.</p>
        </div>
      </div>
    </div>
  </section>
  
  <section class="categories">
    <div class="container">
      <h2 class="section-title">Popular Categories</h2>
      <div class="categories-grid">
        <a href="/products?category=electronics" class="category-card">
          <span class="category-icon">📱</span>
          <h3>Electronics</h3>
        </a>
        
        <a href="/products?category=fashion" class="category-card">
          <span class="category-icon">👕</span>
          <h3>Fashion</h3>
        </a>
        
        <a href="/products?category=home" class="category-card">
          <span class="category-icon">🏠</span>
          <h3>Home & Garden</h3>
        </a>
        
        <a href="/products?category=beauty" class="category-card">
          <span class="category-icon">✨</span>
          <h3>Beauty</h3>
        </a>
        
        <a href="/products?category=sports" class="category-card">
          <span class="category-icon">⚽</span>
          <h3>Sports</h3>
        </a>
        
        <a href="/products?category=toys" class="category-card">
          <span class="category-icon">🧸</span>
          <h3>Toys & Games</h3>
        </a>
      </div>
    </div>
  </section>
</div>

<style lang="scss">
  .home-page {
    width: 100%;
  }
  
  .hero {
    background-color: var(--card-background);
    padding: 4rem 0;
    text-align: center;
    margin-bottom: 3rem;
    border-radius: 0 0 10px 10px;
    
    .hero-content {
      max-width: 800px;
      margin: 0 auto;
      
      h1 {
        font-size: 3rem;
        margin-bottom: 1rem;
        color: var(--primary-color);
      }
      
      p {
        font-size: 1.25rem;
        margin-bottom: 2rem;
        opacity: 0.9;
      }
    }
    
    .cta-buttons {
      display: flex;
      justify-content: center;
      gap: 1rem;
      
      a {
        padding: 0.75rem 1.5rem;
        border-radius: 4px;
        font-weight: 600;
        text-decoration: none;
        transition: transform 0.2s, box-shadow 0.2s;
        
        &:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px var(--shadow-color);
        }
      }
      
      .primary-button {
        background-color: var(--primary-color);
        color: white;
      }
      
      .secondary-button {
        background-color: var(--secondary-color);
        color: white;
      }
    }
  }
  
  .section-title {
    text-align: center;
    margin-bottom: 2.5rem;
    font-size: 2rem;
    color: var(--primary-color);
  }
  
  .features {
    margin-bottom: 4rem;
    
    .features-grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 2rem;
    }
    
    .feature-card {
      background-color: var(--card-background);
      padding: 2rem;
      border-radius: 8px;
      text-align: center;
      box-shadow: 0 4px 6px var(--shadow-color);
      transition: transform 0.3s;
      
      &:hover {
        transform: translateY(-10px);
      }
      
      .feature-icon {
        font-size: 3rem;
        margin-bottom: 1rem;
      }
      
      h3 {
        margin-bottom: 0.5rem;
        color: var(--text-color);
      }
      
      p {
        color: var(--text-color);
        opacity: 0.8;
      }
    }
  }
  
  .categories {
    margin-bottom: 4rem;
    
    .categories-grid {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 1.5rem;
    }
    
    .category-card {
      display: flex;
      flex-direction: column;
      align-items: center;
      padding: 2rem;
      background-color: var(--card-background);
      border-radius: 8px;
      text-decoration: none;
      box-shadow: 0 4px 6px var(--shadow-color);
      transition: transform 0.3s;
      
      &:hover {
        transform: scale(1.05);
      }
      
      .category-icon {
        font-size: 3rem;
        margin-bottom: 1rem;
      }
      
      h3 {
        color: var(--text-color);
        margin: 0;
      }
    }
  }
  
  /* Responsive styles */
  @media (max-width: 1024px) {
    .features .features-grid {
      grid-template-columns: repeat(2, 1fr);
    }
  }
  
  @media (max-width: 768px) {
    .hero .hero-content h1 {
      font-size: 2.5rem;
    }
    
    .categories .categories-grid {
      grid-template-columns: repeat(2, 1fr);
    }
  }
  
  @media (max-width: 576px) {
    .hero .hero-content h1 {
      font-size: 2rem;
    }
    
    .features .features-grid {
      grid-template-columns: 1fr;
    }
    
    .categories .categories-grid {
      grid-template-columns: 1fr;
    }
    
    .hero .cta-buttons {
      flex-direction: column;
    }
  }
</style> 