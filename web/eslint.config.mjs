import js from '@eslint/js';
import nextVitals from 'eslint-config-next/core-web-vitals';
import nextTs from 'eslint-config-next/typescript';
import prettierConfig from 'eslint-config-prettier';
import pluginBetterTailwind from 'eslint-plugin-better-tailwindcss';
import pluginPrettier from 'eslint-plugin-prettier';
import pluginReactHooks from 'eslint-plugin-react-hooks';
import { defineConfig, globalIgnores } from 'eslint/config';
import globals from 'globals';
import tseslint from 'typescript-eslint';

export default defineConfig([
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...nextVitals,
  ...nextTs,

  // Built output, and Go's stock wasm loader that `make web/wasm` copies in.
  globalIgnores(['**/.next/**', '**/out/**', '**/dist/**', '**/next-env.d.ts', '**/wasm_exec.js']),

  {
    plugins: {
      '@typescript-eslint': tseslint.plugin
    },
    languageOptions: {
      parser: tseslint.parser,
      globals: {
        ...globals.browser,
        ...globals.commonjs,
        ...globals.node
      },
      ecmaVersion: 'latest',
      sourceType: 'module',
      parserOptions: {
        ecmaFeatures: {
          jsx: true,
          arrowFunctions: true
        }
      }
    },
    rules: {
      '@typescript-eslint/no-unused-vars': [
        'warn',
        {
          args: 'all',
          argsIgnorePattern: '^_',
          caughtErrors: 'all',
          caughtErrorsIgnorePattern: '^_',
          destructuredArrayIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          ignoreRestSiblings: true
        }
      ],
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-require-imports': 'off',
      '@typescript-eslint/no-empty-function': 'off',
      '@typescript-eslint/no-console': 'off',
      '@next/next/no-img-element': 'off',
      '@next/next/no-html-link-for-pages': 'off',
      // Disable react/* rules from eslint-config-next that have flat config compatibility issues
      'react/display-name': 'off',
      'react/no-direct-mutation-state': 'off',
      'react/no-string-refs': 'off',
      'react/react-in-jsx-scope': 'off',
      'react/prop-types': 'off',
      'react/no-unescaped-entities': 'off',
      'react-hooks/incompatible-library': 'off'
    },
    settings: {
      next: { rootDir: 'apps/site/' },
      react: {
        // Stated rather than 'detect': eslint-plugin-react's detection calls
        // `context.getFilename()`, which ESLint 10 removed, and crashes the run.
        version: '19'
      },
      tailwindcss: {
        callees: ['classNames', 'classname', 'cn', 'clsx', 'ctl', 'cva', 'tv']
      }
    }
  },
  {
    files: ['**/*.{js,jsx,ts,tsx}'],
    plugins: {
      'react-hooks': pluginReactHooks
    },
    rules: {
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'off',
      'react-hooks/set-state-in-effect': 'warn'
    }
  },
  // Tailwind 4 class correctness in the site and the ui package: no class that
  // conflicts with another, no duplicate, no deprecated or unknown utility.
  // Order is prettier-plugin-tailwindcss's job, so the stylistic rules stay off.
  {
    files: ['apps/site/src/**/*.{ts,tsx}', 'packages/ui/src/**/*.{ts,tsx}'],
    plugins: { 'better-tailwindcss': pluginBetterTailwind },
    settings: {
      'better-tailwindcss': {
        entryPoint: 'apps/site/src/styles/main.css',
        callees: ['cn', 'clsx', 'cva', 'tv']
      }
    },
    rules: {
      ...pluginBetterTailwind.configs['correctness-warn'].rules,
      'better-tailwindcss/no-conflicting-classes': 'error'
    }
  },
  {
    files: ['**/*.{js,jsx,ts,tsx,mjs,mts,vue}'],
    plugins: { prettier: pluginPrettier },
    rules: {
      'prettier/prettier': 'warn'
    }
  },
  prettierConfig
]);
