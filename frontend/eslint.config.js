import js from "@eslint/js";
import svelte from "eslint-plugin-svelte";
import globals from "globals";
import ts from "typescript-eslint";
import svelteConfig from "./svelte.config.js";

export default ts.config(
  js.configs.recommended,
  ...ts.configs.recommended,
  ...svelte.configs.recommended,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },
  {
    files: ["**/*.svelte", "**/*.svelte.ts", "**/*.svelte.js"],
    languageOptions: {
      parserOptions: {
        projectService: true,
        extraFileExtensions: [".svelte"],
        parser: ts.parser,
        svelteConfig,
      },
    },
  },
  {
    rules: {
      "svelte/no-trailing-spaces": "error",
      "svelte/spaced-html-comment": "error",
      "svelte/shorthand-attribute": "error",
      "svelte/shorthand-directive": "error",
      "svelte/mustache-spacing": "error",
      "svelte/no-spaces-around-equal-signs-in-attribute": "error",
      "svelte/html-quotes": [
        "error",
        {
          prefer: "double",
          dynamic: {
            quoted: false,
            avoidInvalidUnquotedInHTML: false,
          },
        },
      ],
      "svelte/html-closing-bracket-spacing": [
        "error",
        {
          startTag: "never",
          endTag: "never",
          selfClosingTag: "always",
        },
      ],
      "svelte/prefer-const": [
        "error",
        {
          destructuring: "any",
          excludedRunes: ["$props", "$derived"],
        },
      ],
      "svelte/prefer-destructured-store-props": "error",
      "svelte/button-has-type": [
        "error",
        {
          button: true,
          submit: true,
          reset: true,
        },
      ],
      "svelte/no-inline-styles": [
        "error",
        {
          allowTransitions: true,
        },
      ],
      semi: ["error", "always"],
      quotes: ["error", "double"],
      "@typescript-eslint/no-unused-vars": ["error"],
      "@typescript-eslint/no-explicit-any": ["error"],
      "no-console": "warn",
    },
  }
);
