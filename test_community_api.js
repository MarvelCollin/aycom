// Simple script to test the community API
const form = new FormData();
form.append('name', 'Test Community');
form.append('description', 'Testing community creation');
form.append('rules', 'Test rules');
form.append('categories', JSON.stringify(['News', 'Education']));

// We'll need to replace this with an actual file in the browser
// form.append('icon', new File(['test'], 'icon.png', { type: 'image/png' }));
// form.append('banner', new File(['test'], 'banner.png', { type: 'image/png' }));

console.log('Form data:');
for (const pair of form.entries()) {
  console.log(`${pair[0]}: ${pair[1]}`);
}

// Usage:
// Copy this script into your browser console
// Replace the File objects with actual file input files
// Then send the request manually or use the community.ts createCommunity function 