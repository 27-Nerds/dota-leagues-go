import './app.css';
import { mount } from 'svelte';
import App from './App.svelte';
import { legacyDestination } from './lib/routes.js';

// Old shared hash links redirect to their crawlable page URL.
const destination = legacyDestination(window.location.hash);
if (destination) {
  window.location.replace(destination);
} else {
  const target = document.getElementById('app');
  target.replaceChildren();
  mount(App, { target });
}
