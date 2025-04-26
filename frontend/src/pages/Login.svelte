<script lang="ts">
  let username = '';
  let password = '';
  
  function login() {
    console.log('Login attempted with:', username);
    window.location.href = '/home';
  }
</script>

<div class="min-h-screen bg-gray-900 flex items-center justify-center p-4">
  <div class="max-w-md w-full bg-black rounded-lg p-6 relative">
    <!-- X Logo -->
    <div class="flex justify-center mb-6">
      <span class="text-white text-2xl font-bold">AY</span>
    </div>
    
    <!-- Close Button -->
    <button class="absolute top-3 left-3 text-white" on:click={() => window.location.href='/'}>
      ✕
    </button>
    
    <h2 class="text-white text-xl font-bold mb-6">Sign in to AY</h2>
    
    <!-- Social Login Buttons -->
    <div class="mb-4">
      <button class="w-full py-2 mb-2 bg-white text-black rounded-full flex items-center justify-center font-medium">
        <span class="mr-2">G</span> Sign in with Google
      </button>
      
      <button class="w-full py-2 bg-white text-black rounded-full flex items-center justify-center font-medium">
        <span class="mr-2">🍎</span> Sign in with Apple
      </button>
    </div>
    
    <div class="flex items-center my-4">
      <div class="flex-grow h-px bg-gray-800"></div>
      <div class="mx-4 text-gray-500">or</div>
      <div class="flex-grow h-px bg-gray-800"></div>
    </div>
    
    <!-- Username/Email Field -->
    <div class="mb-4">
      <input 
        type="text" 
        bind:value={username} 
        placeholder="Phone, email, or username"
        class="w-full bg-black border border-gray-800 text-white p-3 rounded-md focus:border-blue-500 focus:outline-none"
      />
    </div>
    
    <!-- Next Button -->
    <button 
      on:click={login} 
      class="w-full py-3 mb-4 bg-white text-black text-center rounded-full font-semibold hover:bg-gray-200 transition-colors"
    >
      Next
    </button>
    
    <!-- Forgot Password Link -->
    <button class="w-full py-3 mb-4 border border-gray-800 text-white text-center rounded-full font-semibold hover:bg-gray-900 transition-colors">
      Forgot password?
    </button>
    
    <!-- Sign Up Link -->
    <p class="text-gray-400 text-sm text-center mt-6">
      Don't have an account? <a href="/register" class="text-blue-500">Sign up</a>
    </p>
  </div>
</div> 