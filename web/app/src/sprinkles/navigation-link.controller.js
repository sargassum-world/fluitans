import { Controller } from 'stimulus';

export default class extends Controller {
  connect() {
    this.updateActiveListener = (event) => {
      const location = event.path[2].location.href;
      if (this.element === undefined) {
        return;
      }

      if (this.element.href === location) {
        this.element.classList.add('is-active');
      } else {
        this.element.classList.remove('is-active');
      }
    };

    document.addEventListener('turbo:render', this.updateActiveListener);
  }
  disconnect() {
    if (this.updateActiveListener === undefined) {
      return;
    }

    document.removeEventListener('turbo:render', this.updateActiveListener);
  }

  updateActive
}
