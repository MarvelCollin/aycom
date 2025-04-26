<script lang="ts">
  import { onMount } from 'svelte';
  
  let step = 1;
  let name = '';
  let email = '';
  let month = '';
  let day = '';
  let year = '';
  let verificationCode = '';
  
  function nextStep() {
    step++;
  }
  
  function goBack() {
    step--;
  }
  
  function submitRegistration() {
    // Handle registration submission
    console.log('Registration submitted');
    window.location.href = '/';
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
    
    {#if step === 1}
    <!-- Step 1: Account Information -->
    <div class="mb-8">
      <h2 class="text-white text-xl font-bold mb-6">Create your account</h2>
      
      <!-- Name Field -->
      <div class="mb-4">
        <label class="block mb-1 text-xs text-gray-400">Name</label>
        <input 
          type="text" 
          bind:value={name} 
          class="w-full bg-black border border-gray-800 text-white p-2 rounded-md focus:border-blue-500 focus:outline-none"
          placeholder="Name"
        />
      </div>
      
      <!-- Email Field -->
      <div class="mb-4">
        <label class="block mb-1 text-xs text-gray-400">Email</label>
        <input 
          type="email" 
          bind:value={email} 
          class="w-full bg-black border border-gray-800 text-white p-2 rounded-md focus:border-blue-500 focus:outline-none"
          placeholder="Email"
        />
      </div>
      
      <!-- Date of Birth -->
      <div class="mb-4">
        <label class="block mb-1 text-xs text-gray-400">Date of birth</label>
        <p class="text-xs text-gray-500 mb-2">This will not be shown publicly. Confirm your own age, even if this account is for a business, a pet, or something else.</p>
        
        <div class="flex gap-2">
          <div class="w-1/2">
            <select 
              bind:value={month} 
              class="w-full bg-black border border-gray-800 text-white p-2 rounded-md focus:border-blue-500 focus:outline-none"
            >
              <option value="" disabled selected>Month</option>
              <option value="01">January</option>
              <option value="02">February</option>
              <option value="03">March</option>
              <!-- Add other months -->
            </select>
          </div>
          
          <div class="w-1/4">
            <select 
              bind:value={day} 
              class="w-full bg-black border border-gray-800 text-white p-2 rounded-md focus:border-blue-500 focus:outline-none"
            >
              <option value="" disabled selected>Day</option>
              <!-- Generate options 1-31 -->
              {#each Array(31) as _, i}
                <option value={i+1}>{i+1}</option>
              {/each}
            </select>
          </div>
          
          <div class="w-1/4">
            <select 
              bind:value={year} 
              class="w-full bg-black border border-gray-800 text-white p-2 rounded-md focus:border-blue-500 focus:outline-none"
            >
              <option value="" disabled selected>Year</option>
              <!-- Generate year options (current year - 100) to current year -->
              {#each Array(100) as _, i}
                <option value={new Date().getFullYear() - i}>{new Date().getFullYear() - i}</option>
              {/each}
            </select>
          </div>
        </div>
      </div>
    </div>
    
    <button 
      on:click={nextStep} 
      class="w-full py-3 bg-white text-black text-center rounded-full font-semibold hover:bg-gray-200 transition-colors"
    >
      Next
    </button>
    
    {/if}
    
    {#if step === 2}
    <!-- Step 2: Verification Code -->
    <div class="mb-8">
      <button on:click={goBack} class="mb-4 text-white flex items-center">
        ← Back
      </button>
      
      <h2 class="text-white text-xl font-bold mb-2">We sent you a code</h2>
      <p class="text-gray-400 text-sm mb-6">Enter it below to verify {email}</p>
      
      <div class="mb-4">
        <label class="block mb-1 text-xs text-gray-400">Verification code</label>
        <input 
          type="text" 
          bind:value={verificationCode} 
          class="w-full bg-black border border-gray-800 text-white p-2 rounded-md focus:border-blue-500 focus:outline-none"
          placeholder="Verification code"
        />
      </div>
      
      <p class="text-blue-500 text-sm mb-6 cursor-pointer">Didn't receive email?</p>
    </div>
    
    <button 
      on:click={submitRegistration} 
      class="w-full py-3 bg-white text-black text-center rounded-full font-semibold hover:bg-gray-200 transition-colors"
    >
      Next
    </button>
    {/if}
  </div>
</div> 