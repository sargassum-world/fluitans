import * as Turbo from '@hotwired/turbo';
import { Application } from 'stimulus';

import ThemeController from './sprinkles/theme.controller';
import NavigationMenuController from './sprinkles/navigation-menu.controller';
import NavigationLinkController from './sprinkles/navigation-link.controller';

Turbo.session.drive = true;

const Stimulus = Application.start();
Stimulus.register('theme', ThemeController);
Stimulus.register('navigation-menu', NavigationMenuController);
Stimulus.register('navigation-link', NavigationLinkController);

export {};
