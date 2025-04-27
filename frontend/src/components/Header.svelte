<script lang="ts">
  import { theme } from '../hooks/useTheme';
  import ThemeToggle from './ThemeToggle.svelte';
  
  let isMenuOpen = false;
  
  function toggleMenu() {
    isMenuOpen = !isMenuOpen;
  }
</script>

<header>
  <div class="container">
    <div class="header-content">
      <div class="logo">
        <a href="/">AYCOM</a>
      </div>
      
      <nav class:active={isMenuOpen}>
        <ul>
          <li><a href="/">Home</a></li>
          <li><a href="/products">Products</a></li>
          <li><a href="/about">About</a></li>
          <li><a href="/contact">Contact</a></li>
        </ul>
      </nav>
      
      <div class="actions">
        <ThemeToggle size="5" className="theme-toggle" />
        
        <div class="auth-buttons">
          <a href="/login" class="login-btn">Login</a>
          <a href="/register" class="register-btn">Register</a>
        </div>
        
        <button class="menu-toggle" on:click={toggleMenu} aria-label="Toggle menu">
          <span class="bar"></span>
          <span class="bar"></span>
          <span class="bar"></span>
        </button>
      </div>
    </div>
  </div>
</header>

<style>
  header {
    background-color: var(--bg-secondary);
    padding: 1rem 0;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    position: sticky;
    top: 0;
    z-index: 100;
  }
  
  .header-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  
  .logo a {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-primary);
    text-decoration: none;
  }
  
  nav ul {
    display: flex;
    list-style: none;
  }
  
  nav ul li {
    margin-left: 1.5rem;
  }
  
  nav ul li a {
    color: var(--text-primary);
    text-decoration: none;
    transition: color 0.2s;
  }
  
  nav ul li a:hover {
    color: var(--color-primary);
  }
  
  .actions {
    display: flex;
    align-items: center;
  }
  
  .theme-toggle {
    margin-right: 1rem;
  }
  
  .auth-buttons {
    display: flex;
    gap: 0.5rem;
  }
  
  .login-btn {
    color: var(--color-primary);
    text-decoration: none;
    padding: 0.5rem 1rem;
  }
  
  .login-btn:hover {
    text-decoration: underline;
  }
  
  .register-btn {
    background-color: var(--color-primary);
    color: white;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    text-decoration: none;
  }
  
  .register-btn:hover {
    opacity: 0.9;
  }
  
  .menu-toggle {
    display: none;
    flex-direction: column;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
  }
  
  .menu-toggle .bar {
    width: 25px;
    height: 3px;
    background-color: var(--text-primary);
    margin: 3px 0;
    border-radius: 3px;
    transition: 0.3s;
  }
  
  /* Responsive styles */
  @media (max-width: 768px) {
    .auth-buttons {
      display: none;
    }
    
    .menu-toggle {
      display: flex;
    }
    
    nav {
      position: fixed;
      top: 60px;
      left: 0;
      right: 0;
      background-color: var(--bg-secondary);
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
      padding: 1rem;
      transform: translateY(-100%);
      opacity: 0;
      transition: transform 0.3s, opacity 0.3s;
    }
    
    nav.active {
      transform: translateY(0);
      opacity: 1;
    }
    
    nav ul {
      flex-direction: column;
    }
    
    nav ul li {
      margin: 0.5rem 0;
    }
  }
</style>