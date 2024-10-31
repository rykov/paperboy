import { computed, set } from '@ember/object';
import { tracked } from '@glimmer/tracking';
import Controller from '@ember/controller';
import { service } from '@ember/service';

var FORMATS = Object.freeze([
  { id: 'html', order: 0 },
  { id: 'text', order: 1 },
  { id: 'raw', order: 2 },
]);

export default class PreviewRenderController extends Controller {
  @service router;

  allFormats = FORMATS;
  @tracked selectedFormat = FORMATS[0];
  queryParams = ['format'];

  @computed('selectedFormat.id')
  get format() {
    return this.selectedFormat.id;
  }
  set format(value) {
    const f = FORMATS.find((f) => f['id'] === value);
    set(this, 'selectedFormat', f || FORMATS[0]);
  }
}
