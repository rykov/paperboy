const isProduction = process.env.EMBER_ENV === 'production';

module.exports = {
  content: ['./app/**/*.{html,hbs,js,ts}', './public/**/*.html'],
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')],

  // Disable JIT in devo. Faster to import all classes.
  ...(!isProduction && { safelist: [{ pattern: /.*/ }] }),

  // Do we need these?
  theme: { extend: {} },
  variants: {},
};
