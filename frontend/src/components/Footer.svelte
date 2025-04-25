<script lang="ts">
  const currentYear = new Date().getFullYear();
</script>

<footer>
  <div class="container">
    <div class="footer-content">
      <div class="footer-section">
        <h3>AYCOM</h3>
        <p>Your trusted online marketplace for all your needs.</p>
      </div>
      
      <div class="footer-section">
        <h3>Quick Links</h3>
        <ul>
          <li><a href="/">Home</a></li>
          <li><a href="/products">Products</a></li>
          <li><a href="/about">About</a></li>
          <li><a href="/contact">Contact</a></li>
        </ul>
      </div>
      
      <div class="footer-section">
        <h3>Legal</h3>
        <ul>
          <li><a href="/privacy">Privacy Policy</a></li>
          <li><a href="/terms">Terms of Service</a></li>
          <li><a href="/returns">Return Policy</a></li>
        </ul>
      </div>
      
      <div class="footer-section">
        <h3>Connect With Us</h3>
        <div class="social-links">
          <a href="#" aria-label="Facebook">FB</a>
          <a href="#" aria-label="Twitter">TW</a>
          <a href="#" aria-label="Instagram">IG</a>
          <a href="#" aria-label="LinkedIn">LI</a>
        </div>
      </div>
    </div>
    
    <div class="footer-bottom">
      <p>&copy; {currentYear} AYCOM. All rights reserved.</p>
    </div>
  </div>
</footer>

<style lang="scss">
  footer {
    background-color: var(--card-background);
    padding: 3rem 0 1rem;
    margin-top: 2rem;
    border-top: 1px solid var(--border-color);
  }
  
  .footer-content {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 2rem;
    margin-bottom: 2rem;
  }
  
  .footer-section {
    h3 {
      margin-bottom: 1rem;
      font-size: 1.2rem;
    }
    
    p {
      color: var(--text-color);
      opacity: 0.8;
    }
    
    ul {
      list-style: none;
      
      li {
        margin-bottom: 0.5rem;
        
        a {
          color: var(--text-color);
          opacity: 0.8;
          text-decoration: none;
          transition: opacity 0.2s;
          
          &:hover {
            opacity: 1;
            color: var(--primary-color);
          }
        }
      }
    }
  }
  
  .social-links {
    display: flex;
    gap: 1rem;
    margin-top: 0.5rem;
    
    a {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 36px;
      height: 36px;
      background-color: var(--primary-color);
      color: white;
      border-radius: 50%;
      transition: transform 0.2s;
      
      &:hover {
        transform: translateY(-3px);
      }
    }
  }
  
  .footer-bottom {
    padding-top: 1.5rem;
    border-top: 1px solid var(--border-color);
    text-align: center;
    
    p {
      color: var(--text-color);
      opacity: 0.7;
      font-size: 0.9rem;
    }
  }
  
  /* Responsive styles */
  @media (max-width: 1024px) {
    .footer-content {
      grid-template-columns: repeat(2, 1fr);
    }
  }
  
  @media (max-width: 576px) {
    .footer-content {
      grid-template-columns: 1fr;
    }
  }
</style> 