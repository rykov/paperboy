import { helper } from '@ember/component/helper';

export function contains([value, sub /*, ...rest */]) {
  return value.indexOf(sub) >= 0;
}

export default helper(contains);
