const isProduction = process.env.EMBER_ENV === 'production';

module.exports = {
  mode: isProduction ? 'jit' : undefined,
  purge: ['./app/**/*.{html,hbs,js,ts}', './public/**/*.html'],
  theme: {
    extend: {},
  },
  variants: {},
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')],
};
