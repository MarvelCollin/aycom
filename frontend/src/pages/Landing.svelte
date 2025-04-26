<script lang="ts">
  // Landing page for X (Twitter) clone
</script>

<div class="flex min-h-screen bg-black text-white">
  <div class="container mx-auto flex flex-col md:flex-row items-center px-6 py-8">
    <!-- Logo Section - Left side -->
    <div class="w-full md:w-1/2 flex justify-center md:justify-start mb-8 md:mb-0">
      <div class="w-52 h-52 relative">
        <span class="text-white text-9xl font-bold">AY</span>
      </div>
    </div>
    
    <!-- Content Section - Right side -->
    <div class="w-full md:w-1/2 flex flex-col md:pl-8">
      <h1 class="text-6xl font-bold mb-6">Happening now</h1>
      <h2 class="text-3xl font-bold mb-4">Join today.</h2>
      <p class="text-xl mb-8">Connect, share, engage.</p>
      
      <!-- Account Creation Section -->
      <div class="w-full max-w-sm">
        <div class="mb-6">
          <p class="text-sm mb-2">Don't have an account?</p>
          <a href="/register" class="block w-full py-2 bg-blue-500 text-white text-center rounded-full font-semibold hover:bg-blue-600 transition-colors">
            Create account
          </a>
          <p class="text-xs mt-2 text-gray-500">
            By signing up, you agree to the <a href="#" class="text-blue-500">Terms of Service</a> and <a href="#" class="text-blue-500">Privacy Policy</a>, including <a href="#" class="text-blue-500">Cookie Use</a>.
          </p>
        </div>
        
        <div class="pt-4 mt-4 border-t border-gray-800">
          <p class="text-sm mb-2">Already have an account?</p>
          <a href="/login" class="block w-full py-2 border border-gray-800 text-center rounded-full font-semibold hover:bg-gray-900 transition-colors">
            <span class="text-blue-500">Sign in</span>
          </a>
        </div>
      </div>
    </div>
  </div>
</div> 