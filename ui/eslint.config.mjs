import js from '@eslint/js'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'
import jsxA11y from 'eslint-plugin-jsx-a11y'
import importX from 'eslint-plugin-import-x'
import testingLibrary from 'eslint-plugin-testing-library'
import tseslint from 'typescript-eslint'

const files = ['src/**/*.{ts,tsx}']

export default tseslint.config(
  {
    ignores: ['dist', 'coverage', 'eslint.config.mjs', 'vite.config.ts']
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...tseslint.configs.stylistic,
  {
    files,
    languageOptions: {
      parserOptions: {
        project: undefined
      }
    },
    plugins: {
      react,
      'react-hooks': reactHooks,
      'jsx-a11y': jsxA11y,
      import: importX,
      'testing-library': testingLibrary
    },
    settings: {
      react: {
        version: 'detect'
      }
    },
    rules: {
      ...react.configs.recommended.rules,
      ...react.configs['jsx-runtime'].rules,
      ...jsxA11y.configs.recommended.rules,
      '@typescript-eslint/consistent-type-definitions': 'off',
      'react/jsx-uses-react': 'off',
      'react/react-in-jsx-scope': 'off',
      'react/prop-types': 'off',
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
      'import/order': [
        'warn',
        {
          groups: ['builtin', 'external', 'internal', 'parent', 'sibling', 'index', 'object'],
          'newlines-between': 'always'
        }
      ]
    }
  },
  {
    files: ['src/test/**/*.{ts,tsx}'],
    plugins: {
      'testing-library': testingLibrary
    },
    rules: {
      ...testingLibrary.configs.react.rules
    }
  }
)
