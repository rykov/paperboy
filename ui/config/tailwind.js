const isProduction = process.env.EMBER_ENV === 'production';

module.exports = {
  purge: ['./app/**/*.{html,hbs,js,ts}', './public/**/*.html'],
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')],

  // Disable JIT in devo. Faster to import all classes.
  safelist: isProduction ? undefined : [{ pattern: /.*/ }],

  // Do we need these?
  theme: { extend: {} },
  variants: {},
};
