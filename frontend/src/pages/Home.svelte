<script lang="ts">
  // Home page component
</script>

<div class="w-full">
  <section class="bg-card rounded-b-lg py-16 mb-12 text-center">
    <div class="container mx-auto">
      <div class="max-w-3xl mx-auto">
        <h1 class="text-5xl mb-4 text-primary font-bold">Welcome to AYCOM</h1>
        <p class="text-xl mb-8 opacity-90">Your trusted online marketplace for all your needs</p>
        <div class="flex justify-center gap-4">
          <a href="/products" class="py-3 px-6 rounded-md font-semibold bg-primary text-white transition-transform hover:translate-y-[-2px] shadow-md">Browse Products</a>
          <a href="/register" class="py-3 px-6 rounded-md font-semibold bg-secondary text-white transition-transform hover:translate-y-[-2px] shadow-md">Join Now</a>
        </div>
      </div>
    </div>
  </section>
  
  <section class="mb-16">
    <div class="container mx-auto">
      <h2 class="text-4xl font-bold text-center mb-10 text-primary">Why Choose AYCOM?</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
        <div class="bg-card p-8 rounded-lg text-center shadow-md transition-transform hover:translate-y-[-10px]">
          <div class="text-5xl mb-4">🚚</div>
          <h3 class="font-bold text-xl mb-2">Fast Delivery</h3>
          <p class="opacity-80">Get your products delivered to your doorstep in no time.</p>
        </div>
        
        <div class="bg-card p-8 rounded-lg text-center shadow-md transition-transform hover:translate-y-[-10px]">
          <div class="text-5xl mb-4">💎</div>
          <h3 class="font-bold text-xl mb-2">Quality Products</h3>
          <p class="opacity-80">All our products are carefully selected for quality.</p>
        </div>
        
        <div class="bg-card p-8 rounded-lg text-center shadow-md transition-transform hover:translate-y-[-10px]">
          <div class="text-5xl mb-4">🔒</div>
          <h3 class="font-bold text-xl mb-2">Secure Payments</h3>
          <p class="opacity-80">Shop with confidence with our secure payment options.</p>
        </div>
        
        <div class="bg-card p-8 rounded-lg text-center shadow-md transition-transform hover:translate-y-[-10px]">
          <div class="text-5xl mb-4">🌍</div>
          <h3 class="font-bold text-xl mb-2">Global Reach</h3>
          <p class="opacity-80">Connect with sellers and buyers from around the world.</p>
        </div>
      </div>
    </div>
  </section>
  
  <section class="mb-16">
    <div class="container mx-auto">
      <h2 class="text-4xl font-bold text-center mb-10 text-primary">Popular Categories</h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        <a href="/products?category=electronics" class="flex flex-col items-center p-8 bg-card rounded-lg shadow-md transition-transform hover:scale-105 text-center">
          <span class="text-5xl mb-4">📱</span>
          <h3 class="font-bold text-xl">Electronics</h3>
        </a>
        
        <a href="/products?category=fashion" class="flex flex-col items-center p-8 bg-card rounded-lg shadow-md transition-transform hover:scale-105 text-center">
          <span class="text-5xl mb-4">👕</span>
          <h3 class="font-bold text-xl">Fashion</h3>
        </a>
        
        <a href="/products?category=home" class="flex flex-col items-center p-8 bg-card rounded-lg shadow-md transition-transform hover:scale-105 text-center">
          <span class="text-5xl mb-4">🏠</span>
          <h3 class="font-bold text-xl">Home & Garden</h3>
        </a>
        
        <a href="/products?category=beauty" class="flex flex-col items-center p-8 bg-card rounded-lg shadow-md transition-transform hover:scale-105 text-center">
          <span class="text-5xl mb-4">✨</span>
          <h3 class="font-bold text-xl">Beauty</h3>
        </a>
        
        <a href="/products?category=sports" class="flex flex-col items-center p-8 bg-card rounded-lg shadow-md transition-transform hover:scale-105 text-center">
          <span class="text-5xl mb-4">⚽</span>
          <h3 class="font-bold text-xl">Sports</h3>
        </a>
        
        <a href="/products?category=toys" class="flex flex-col items-center p-8 bg-card rounded-lg shadow-md transition-transform hover:scale-105 text-center">
          <span class="text-5xl mb-4">🧸</span>
          <h3 class="font-bold text-xl">Toys & Games</h3>
        </a>
      </div>
    </div>
  </section>
</div> 