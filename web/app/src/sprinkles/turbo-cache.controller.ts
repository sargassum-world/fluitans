import * as Turbo from '@hotwired/turbo';
import { Controller } from 'stimulus';

export default class extends Controller {
  clear(): void {
    Turbo.clearCache();
  }
}
