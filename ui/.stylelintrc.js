'use strict';

module.exports = {
  extends: ['stylelint-config-standard'],
  customSyntax: 'postcss-scss',
  plugins: ['stylelint-scss'],
  rules: {
    // Support ":global" for CSS-Modules
    'selector-pseudo-class-no-unknown': [true, { ignorePseudoClasses: ['global'] }],

    // Support for SCSS, Tailwind, etc
    'at-rule-no-unknown': [true, { ignoreAtRules: ['tailwind'] }],
    'import-notation': [null], // To allow SCSS imports
  },
};
