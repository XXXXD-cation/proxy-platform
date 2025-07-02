import { globalIgnores } from 'eslint/config'
import { defineConfigWithVueTs, vueTsConfigs } from '@vue/eslint-config-typescript'
import pluginVue from 'eslint-plugin-vue'
import pluginVitest from '@vitest/eslint-plugin'
import skipFormatting from '@vue/eslint-config-prettier/skip-formatting'

// To allow more languages other than `ts` in `.vue` files, uncomment the following lines:
// import { configureVueProject } from '@vue/eslint-config-typescript'
// configureVueProject({ scriptLangs: ['ts', 'tsx'] })
// More info at https://github.com/vuejs/eslint-config-typescript/#advanced-setup

export default defineConfigWithVueTs(
  {
    name: 'app/files-to-lint',
    files: ['**/*.{ts,mts,tsx,vue}'],
  },

  globalIgnores(['**/dist/**', '**/dist-ssr/**', '**/coverage/**', '**/node_modules/**']),

  // Vue基础配置 - 使用推荐规则
  pluginVue.configs['flat/recommended'],
  vueTsConfigs.recommended,
  
  // Vitest测试文件配置
  {
    ...pluginVitest.configs.recommended,
    files: ['src/**/__tests__/*', 'src/**/*.test.*', 'src/**/*.spec.*'],
  },

  // 自定义规则配置
  {
    name: 'app/custom-rules',
    rules: {
      // Vue组件命名规范
      'vue/component-name-in-template-casing': ['error', 'kebab-case'],
      'vue/component-definition-name-casing': ['error', 'PascalCase'],
      
      // HTML属性和指令规范
      'vue/html-self-closing': ['error', {
        html: {
          void: 'never',
          normal: 'always',
          component: 'always',
        },
        svg: 'always',
        math: 'always',
      }],
      'vue/max-attributes-per-line': ['error', {
        singleline: { max: 3 },
        multiline: { max: 1 },
      }],
      
      // Props和事件规范
      'vue/prop-name-casing': ['error', 'camelCase'],
      
      // 代码质量规则
      'vue/no-unused-components': 'error',
      'vue/require-default-prop': 'error',
      'vue/require-prop-types': 'error',
      'vue/no-duplicate-attributes': 'error',
      'vue/no-multiple-template-root': 'off', // Vue 3允许多根节点
      
      // TypeScript相关规则
      '@typescript-eslint/no-unused-vars': ['error', { 
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_',
      }],
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/explicit-module-boundary-types': 'off',
      '@typescript-eslint/no-explicit-any': 'warn',
      
      // 通用代码质量规则
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      'no-debugger': 'error',
      'no-alert': 'warn',
      'prefer-const': 'error',
      'no-var': 'error',
      'object-shorthand': 'error',
      'prefer-template': 'error',
      
      // 命名规范
      camelcase: ['error', { 
        properties: 'never',
        ignoreDestructuring: true,
        ignoreImports: true,
      }],
    },
  },

  // Prettier配置必须放在最后
  skipFormatting,
)
