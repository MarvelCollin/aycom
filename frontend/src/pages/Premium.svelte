<script lang="ts">
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useTheme } from '../hooks/useTheme';
  import CheckIcon from 'svelte-feather-icons/src/icons/CheckIcon.svelte';
  import StarIcon from 'svelte-feather-icons/src/icons/StarIcon.svelte';
  
  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';
  
  const features = [
    { id: 1, text: "Verified profile badge" },
    { id: 2, text: "Priority in search results" },
    { id: 3, text: "Extended character limit in posts" },
    { id: 4, text: "Advanced analytics for your posts" },
    { id: 5, text: "Ad-free experience" },
    { id: 6, text: "Exclusive access to new features" },
  ];
  
  const plans = [
    {
      id: 'basic',
      name: 'Premium Basic',
      price: '$4.99',
      period: 'monthly',
      features: [1, 2, 3]
    },
    {
      id: 'plus',
      name: 'Premium Plus',
      price: '$9.99',
      period: 'monthly',
      features: [1, 2, 3, 4, 5]
    },
    {
      id: 'pro',
      name: 'Premium Pro',
      price: '$14.99',
      period: 'monthly',
      features: [1, 2, 3, 4, 5, 6]
    }
  ];
  
  function subscribe(planId: string) {
    alert(`Subscription to ${planId} plan - This feature is coming soon!`);
  }
</script>

<MainLayout>
  <div class="page-header {isDarkMode ? 'page-header-dark' : ''}">
    <h1 class="page-title">Premium</h1>
  </div>
  
  <div class="premium-container {isDarkMode ? 'premium-container-dark' : ''}">
    <div class="premium-header">
      <div class="premium-icon">
        <StarIcon size="24" />
      </div>
      <h2>AYCOM Premium</h2>
      <p>Unlock enhanced features and stand out from the crowd</p>
    </div>
    
    <div class="plans-container">
      {#each plans as plan}
        <div class="plan-card {isDarkMode ? 'plan-card-dark' : ''}">
          <div class="plan-header">
            <h3>{plan.name}</h3>
            <div class="plan-price">
              <span class="price">{plan.price}</span>
              <span class="period">/{plan.period}</span>
            </div>
          </div>
          
          <ul class="plan-features">
            {#each features as feature}
              {#if plan.features.includes(feature.id)}
                <li class="feature-included">
                  <span class="feature-icon">
                    <CheckIcon size="16" />
                  </span>
                  <span>{feature.text}</span>
                </li>
              {:else}
                <li class="feature-excluded">
                  <span>{feature.text}</span>
                </li>
              {/if}
            {/each}
          </ul>
          
          <button 
            class="subscribe-btn {isDarkMode ? 'subscribe-btn-dark' : ''}"
            on:click={() => subscribe(plan.id)}
          >
            Subscribe
          </button>
        </div>
      {/each}
    </div>
    
    <div class="premium-footer">
      <p>All plans include access to basic AYCOM Premium features.</p>
      <p>Cancel anytime. Prorated refunds are available.</p>
    </div>
  </div>
</MainLayout>

<style>
  .premium-container {
    padding: var(--space-4);
  }
  
  .premium-header {
    text-align: center;
    margin-bottom: var(--space-6);
  }
  
  .premium-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 48px;
    height: 48px;
    background-color: var(--color-primary);
    color: white;
    border-radius: var(--radius-full);
    margin-bottom: var(--space-3);
  }
  
  .premium-header h2 {
    font-size: var(--font-size-2xl);
    margin-bottom: var(--space-2);
    font-weight: var(--font-weight-bold);
  }
  
  .premium-header p {
    color: var(--text-secondary);
    font-size: var(--font-size-lg);
  }
  
  .plans-container {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: var(--space-4);
    margin-bottom: var(--space-6);
  }
  
  .plan-card {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
    transition: transform 0.2s, box-shadow 0.2s;
  }
  
  .plan-card-dark {
    background-color: var(--dark-bg-secondary);
  }
  
  .plan-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  }
  
  .plan-header {
    margin-bottom: var(--space-4);
    text-align: center;
  }
  
  .plan-header h3 {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-2);
  }
  
  .plan-price {
    display: flex;
    align-items: flex-end;
    justify-content: center;
  }
  
  .price {
    font-size: var(--font-size-2xl);
    font-weight: var(--font-weight-bold);
    color: var(--text-primary);
  }
  
  .period {
    font-size: var(--font-size-sm);
    color: var(--text-tertiary);
    margin-left: var(--space-1);
    margin-bottom: 4px;
  }
  
  .plan-features {
    list-style: none;
    padding: 0;
    margin: 0 0 var(--space-4) 0;
  }
  
  .plan-features li {
    padding: var(--space-2) 0;
    display: flex;
    align-items: center;
  }
  
  .feature-included {
    color: var(--text-primary);
  }
  
  .feature-excluded {
    color: var(--text-tertiary);
    text-decoration: line-through;
  }
  
  .feature-icon {
    color: var(--color-success);
    margin-right: var(--space-2);
    display: flex;
    align-items: center;
  }
  
  .subscribe-btn {
    width: 100%;
    padding: var(--space-3);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-full);
    font-weight: var(--font-weight-bold);
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .subscribe-btn:hover {
    background-color: var(--color-primary-hover);
  }
  
  .premium-footer {
    text-align: center;
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
  }
  
  .premium-footer p {
    margin: var(--space-1) 0;
  }
  
  @media (max-width: 768px) {
    .plans-container {
      grid-template-columns: 1fr;
      gap: var(--space-4);
    }
  }
</style> 