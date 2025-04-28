# AYCOM Frontend

Frontend for the AYCOM social media platform.

## Project Structure

The project has been organized into the following structure:

### Components

- **common/**: Reusable UI components across the application
  - `Logo.svelte`: Logo component with theme support
  - `ThemeToggle.svelte`: Button to toggle between dark and light themes
  - `Tweet.svelte`: Component for displaying tweets

- **layout/**: Layout components
  - `AuthLayout.svelte`: Common layout for authentication pages

- **navigation/**: Navigation-related components
  - `Sidebar.svelte`: Left sidebar navigation for the app
  - `RightSidebar.svelte`: Right sidebar with search, trends, etc.

- **forms/**: Form-related components
  - `FormInput.svelte`: Reusable form input component
  - `TweetComposer.svelte`: Component for composing new tweets

- **auth/**: Authentication-related components

### Hooks

- `useTheme.ts`: Hook for managing theme state (dark/light)
- `useAuth.ts`: Hook for authentication operations
- `useValidation.ts`: Hook for form validation functions
- `useExternalServices.ts`: Hook for external services like Google Auth
- `useTweets.ts`: Hook for managing tweets with CRUD operations

## Pages

- `Landing.svelte`: Landing page
- `Login.svelte`: Login page
- `Register.svelte`: Registration page
- `Feed.svelte`: Main feed page with tweets
- `GoogleCallback.svelte`: Callback handler for Google authentication
- `Home.svelte`: Home page with marketing content

## Development

1. Install dependencies:
```bash
npm install
```

2. Run the development server:
```bash
npm run dev
```

3. Build for production:
```bash
npm run build
```

# Svelte + TS + Vite

This template should help get you started developing with Svelte and TypeScript in Vite.

## Technical considerations

**Why use this over SvelteKit?**

- It brings its own routing solution which might not be preferable for some users.
- It is first and foremost a framework that just happens to use Vite under the hood, not a Vite app.

This template contains as little as possible to get started with Vite + TypeScript + Svelte, while taking into account the developer experience with regards to HMR and intellisense. It demonstrates capabilities on par with the other `create-vite` templates and is a good starting point for beginners dipping their toes into a Vite + Svelte project.

Should you later need the extended capabilities and extensibility provided by SvelteKit, the template has been structured similarly to SvelteKit so that it is easy to migrate.

**Why `global.d.ts` instead of `compilerOptions.types` inside `jsconfig.json` or `tsconfig.json`?**

Setting `compilerOptions.types` shuts out all other types not explicitly listed in the configuration. Using triple-slash references keeps the default TypeScript setting of accepting type information from the entire workspace, while also adding `svelte` and `vite/client` type information.

**Why include `.vscode/extensions.json`?**

Other templates indirectly recommend extensions via the README, but this file allows VS Code to prompt the user to install the recommended extension upon opening the project.

**Why enable `allowJs` in the TS template?**

While `allowJs: false` would indeed prevent the use of `.js` files in the project, it does not prevent the use of JavaScript syntax in `.svelte` files. In addition, it would force `checkJs: false`, bringing the worst of both worlds: not being able to guarantee the entire codebase is TypeScript, and also having worse typechecking for the existing JavaScript. In addition, there are valid use cases in which a mixed codebase may be relevant.

**Why is HMR not preserving my local component state?**

HMR state preservation comes with a number of gotchas! It has been disabled by default in both `svelte-hmr` and `@sveltejs/vite-plugin-svelte` due to its often surprising behavior. You can read the details [here](https://github.com/rixo/svelte-hmr#svelte-hmr).

If you have state that's important to retain within a component, consider creating an external store which would not be replaced by HMR.

```ts
// store.ts
// An extremely simple external store
import { writable } from 'svelte/store'
export default writable(0)
```
