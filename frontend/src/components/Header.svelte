<script lang="ts">
  export let theme: 'light' | 'dark';
  export let toggleTheme: () => void;
  
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
        <button class="theme-toggle" on:click={toggleTheme} aria-label="Toggle theme">
          {#if theme === 'light'}
            <span>🌙</span>
          {:else}
            <span>☀️</span>
          {/if}
        </button>
        
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

<style lang="scss">
  header {
    background-color: var(--card-background);
    padding: 1rem 0;
    box-shadow: 0 2px 4px var(--shadow-color);
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
    color: var(--primary-color);
    text-decoration: none;
  }
  
  nav {
    ul {
      display: flex;
      list-style: none;
      
      li {
        margin-left: 1.5rem;
        
        a {
          color: var(--text-color);
          text-decoration: none;
          transition: color 0.2s;
          
          &:hover {
            color: var(--primary-color);
          }
        }
      }
    }
  }
  
  .actions {
    display: flex;
    align-items: center;
  }
  
  .theme-toggle {
    background: none;
    border: none;
    font-size: 1.25rem;
    margin-right: 1rem;
    padding: 0.25rem;
    cursor: pointer;
    
    span {
      display: block;
    }
  }
  
  .auth-buttons {
    display: flex;
    gap: 0.5rem;
    
    .login-btn {
      color: var(--primary-color);
      text-decoration: none;
      padding: 0.5rem 1rem;
      
      &:hover {
        text-decoration: underline;
      }
    }
    
    .register-btn {
      background-color: var(--primary-color);
      color: white;
      padding: 0.5rem 1rem;
      border-radius: 4px;
      text-decoration: none;
      
      &:hover {
        background-color: darken(#4f46e5, 10%);
      }
    }
  }
  
  .menu-toggle {
    display: none;
    flex-direction: column;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
    
    .bar {
      width: 25px;
      height: 3px;
      background-color: var(--text-color);
      margin: 3px 0;
      border-radius: 3px;
      transition: 0.3s;
    }
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
      background-color: var(--card-background);
      box-shadow: 0 4px 8px var(--shadow-color);
      padding: 1rem;
      transform: translateY(-100%);
      opacity: 0;
      transition: transform 0.3s, opacity 0.3s;
      
      &.active {
        transform: translateY(0);
        opacity: 1;
      }
      
      ul {
        flex-direction: column;
        
        li {
          margin: 0.5rem 0;
        }
      }
    }
  }
</style> 