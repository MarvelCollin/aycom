<script lang="ts">
  import { useTheme } from '../hooks/useTheme';
  import { onMount } from 'svelte';
  
  // Get theme store
  const { theme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  onMount(() => {
    // Apply theme class to document when component mounts
    document.documentElement.classList.add(isDarkMode ? 'dark' : 'light');
  });
</script>

<div class="w-full min-h-screen {isDarkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'} transition-colors duration-300">
  <section class="py-20 mb-12 text-center relative overflow-hidden {isDarkMode ? 'bg-gray-800' : 'bg-white'} shadow-md">
    <div class="absolute inset-0 z-0 {isDarkMode ? 'opacity-10' : 'opacity-5'}">
      <div class="absolute inset-0 bg-gradient-to-r from-blue-500 to-purple-600 transform -rotate-6 scale-110 translate-x-1/4 translate-y-1/4 rounded-full blur-3xl"></div>
    </div>
    <div class="container mx-auto relative z-10">
      <div class="max-w-3xl mx-auto px-6">
        <h1 class="text-4xl md:text-6xl mb-6 font-bold {isDarkMode ? 'text-white' : 'text-gray-900'}">Welcome to <span class="text-blue-500">AYCOM</span></h1>
        <p class="text-xl mb-10 {isDarkMode ? 'text-gray-300' : 'text-gray-600'}">Your trusted online marketplace for all your needs</p>
        <div class="flex flex-col sm:flex-row justify-center gap-4">
          <a href="/products" class="py-3 px-8 rounded-full font-semibold bg-blue-500 text-white hover:bg-blue-600 transition-all hover:shadow-lg transform hover:-translate-y-1">
            Browse Products
          </a>
          <a href="/register" class="py-3 px-8 rounded-full font-semibold {isDarkMode ? 'bg-gray-700 hover:bg-gray-600' : 'bg-gray-200 hover:bg-gray-300'} transition-all hover:shadow-lg transform hover:-translate-y-1">
            Join Now
          </a>
        </div>
      </div>
    </div>
  </section>
  
  <section class="mb-20">
    <div class="container mx-auto px-6">
      <h2 class="text-3xl md:text-4xl font-bold text-center mb-12 text-blue-500">Why Choose AYCOM?</h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-8">
        <div class="{isDarkMode ? 'bg-gray-800' : 'bg-white'} p-8 rounded-xl text-center shadow-lg transition-transform hover:-translate-y-2 duration-300">
          <div class="text-5xl mb-6 mx-auto bg-blue-500 bg-opacity-10 rounded-full w-20 h-20 flex items-center justify-center">🚚</div>
          <h3 class="font-bold text-xl mb-3 {isDarkMode ? 'text-white' : 'text-gray-900'}">Fast Delivery</h3>
          <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-600'}">Get your products delivered to your doorstep in no time.</p>
        </div>
        
        <div class="{isDarkMode ? 'bg-gray-800' : 'bg-white'} p-8 rounded-xl text-center shadow-lg transition-transform hover:-translate-y-2 duration-300">
          <div class="text-5xl mb-6 mx-auto bg-blue-500 bg-opacity-10 rounded-full w-20 h-20 flex items-center justify-center">💎</div>
          <h3 class="font-bold text-xl mb-3 {isDarkMode ? 'text-white' : 'text-gray-900'}">Quality Products</h3>
          <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-600'}">All our products are carefully selected for quality.</p>
        </div>
        
        <div class="{isDarkMode ? 'bg-gray-800' : 'bg-white'} p-8 rounded-xl text-center shadow-lg transition-transform hover:-translate-y-2 duration-300">
          <div class="text-5xl mb-6 mx-auto bg-blue-500 bg-opacity-10 rounded-full w-20 h-20 flex items-center justify-center">🔒</div>
          <h3 class="font-bold text-xl mb-3 {isDarkMode ? 'text-white' : 'text-gray-900'}">Secure Payments</h3>
          <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-600'}">Shop with confidence with our secure payment options.</p>
        </div>
        
        <div class="{isDarkMode ? 'bg-gray-800' : 'bg-white'} p-8 rounded-xl text-center shadow-lg transition-transform hover:-translate-y-2 duration-300">
          <div class="text-5xl mb-6 mx-auto bg-blue-500 bg-opacity-10 rounded-full w-20 h-20 flex items-center justify-center">🌍</div>
          <h3 class="font-bold text-xl mb-3 {isDarkMode ? 'text-white' : 'text-gray-900'}">Global Reach</h3>
          <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-600'}">Connect with sellers and buyers from around the world.</p>
        </div>
      </div>
    </div>
  </section>
  
  <section class="mb-20">
    <div class="container mx-auto px-6">
      <h2 class="text-3xl md:text-4xl font-bold text-center mb-12 text-blue-500">Popular Categories</h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        <a href="/products?category=electronics" class="flex flex-col items-center p-8 {isDarkMode ? 'bg-gray-800 hover:bg-gray-700' : 'bg-white hover:bg-gray-50'} rounded-xl shadow-lg transition-all hover:-translate-y-1 hover:shadow-xl text-center">
          <div class="text-5xl mb-6 mx-auto bg-blue-500 bg-opacity-10 rounded-full w-20 h-20 flex items-center justify-center">📱</div>
          <h3 class="font-bold text-xl {isDarkMode ? 'text-white' : 'text-gray-900'}">Electronics</h3>
        </a>
        
        <a href="/products?category=fashion" class="flex flex-col items-center p-8 {isDarkMode ? 'bg-gray-800 hover:bg-gray-700' : 'bg-white hover:bg-gray-50'} rounded-xl shadow-lg transition-all hover:-translate-y-1 hover:shadow-xl text-center">
          <div class="text-5xl mb-6 mx-auto bg-blue-500 bg-opacity-10 rounded-full w-20 h-20 flex items-center justify-center">👕</div>
          <h3 class="font-bold text-xl {isDarkMode ? 'text-white' : 'text-gray-900'}">Fashion</h3>
        </a>
        
        <a href="/products?category=home" class="flex flex-col items-center p-8 {isDarkMode ? 'bg-gray-800 hover:bg-gray-700' : 'bg-white hover:bg-gray-50'} rounded-xl shadow-lg transition-all hover:-translate-y-1 hover:shadow-xl text-center">
          <div class="text-5xl mb-6 mx-auto bg-blue-500 bg-opacity-10 rounded-full w-20 h-20 flex items-center justify-center">🏠</div>
          <h3 class="font-bold text-xl {isDarkMode ? 'text-white' : 'text-gray-900'}">Home & Garden</h3>
        </a>
        
        <a href="/products?category=beauty" class="flex flex-col items-center p-8 {isDarkMode ? 'bg-gray-800 hover:bg-gray-700' : 'bg-white hover:bg-gray-50'} rounded-xl shadow-lg transition-all hover:-translate-y-1 hover:shadow-xl text-center">
          <div class="text-5xl mb-6 mx-auto bg-blue-500 bg-opacity-10 rounded-full w-20 h-20 flex items-center justify-center">✨</div>
          <h3 class="font-bold text-xl {isDarkMode ? 'text-white' : 'text-gray-900'}">Beauty</h3>
        </a>
        
        <a href="/products?category=sports" class="flex flex-col items-center p-8 {isDarkMode ? 'bg-gray-800 hover:bg-gray-700' : 'bg-white hover:bg-gray-50'} rounded-xl shadow-lg transition-all hover:-translate-y-1 hover:shadow-xl text-center">
          <div class="text-5xl mb-6 mx-auto bg-blue-500 bg-opacity-10 rounded-full w-20 h-20 flex items-center justify-center">⚽</div>
          <h3 class="font-bold text-xl {isDarkMode ? 'text-white' : 'text-gray-900'}">Sports</h3>
        </a>
        
        <a href="/products?category=toys" class="flex flex-col items-center p-8 {isDarkMode ? 'bg-gray-800 hover:bg-gray-700' : 'bg-white hover:bg-gray-50'} rounded-xl shadow-lg transition-all hover:-translate-y-1 hover:shadow-xl text-center">
          <div class="text-5xl mb-6 mx-auto bg-blue-500 bg-opacity-10 rounded-full w-20 h-20 flex items-center justify-center">🧸</div>
          <h3 class="font-bold text-xl {isDarkMode ? 'text-white' : 'text-gray-900'}">Toys & Games</h3>
        </a>
      </div>
    </div>
  </section>
  
  <section class="pb-20">
    <div class="container mx-auto px-6 text-center">
      <h2 class="text-3xl md:text-4xl font-bold mb-8 text-blue-500">Ready to start shopping?</h2>
      <a href="/register" class="inline-block py-3 px-8 rounded-full font-semibold bg-blue-500 text-white hover:bg-blue-600 transition-all hover:shadow-lg transform hover:-translate-y-1">
        Create Account
      </a>
    </div>
  </section>
</div> 