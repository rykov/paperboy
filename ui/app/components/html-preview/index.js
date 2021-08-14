import { inject as service } from '@ember/service';
import { tracked } from '@glimmer/tracking';
import Component from '@glimmer/component';

export default class HtmlPreviewComponent extends Component {
  @service resize;

  @tracked height = 500;
  _iframe = null;

  constructor() {
    super(...arguments);
    this.resize.on('didResize', () => this.didResize());
  }

  didInsert(element, [instance]) {
    instance._iframe = element;
    instance.didResize();
  }

  didResize() {
    const rect = this._iframe?.getBoundingClientRect();
    this.height = window.innerHeight - (rect?.top || 0);
  }
}
