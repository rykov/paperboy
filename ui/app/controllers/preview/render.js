import { inject as service } from '@ember/service';
import { computed, set } from '@ember/object';
import { tracked } from '@glimmer/tracking';
import Controller from '@ember/controller';

var FORMATS = [
  { id: 'html', order: 0 },
  { id: 'text', order: 1 },
  { id: 'raw', order: 2 },
];

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
    var f = FORMATS.findBy('id', value) || FORMATS[0];
    set(this, 'selectedFormat', f);
  }
}
