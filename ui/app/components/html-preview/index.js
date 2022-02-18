import { inject as service } from '@ember/service';
import { tracked } from '@glimmer/tracking';
import Component from '@glimmer/component';
import { action } from '@ember/object';

export default class HtmlPreviewComponent extends Component {
  @service resizeObserver;

  @tracked height = 500;
  _iframe = null;

  constructor() {
    super(...arguments);
  }

  willDestroy() {
    super.willDestroy(...arguments);
    if (this._iframe) {
      this.resizeObserver.unobserve(document.body, this.didResize);
    }
  }

  didInsert(element, [instance]) {
    instance.resizeObserver.observe(document.body, instance.didResize);
    instance._iframe = element;
    instance.didResize();
  }

  // Action to self-bind
  @action didResize() {
    const rect = this._iframe?.getBoundingClientRect();
    this.height = window.innerHeight - (rect?.top || 0);
  }
}
