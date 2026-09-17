import { writable } from 'svelte/store';

function createAuthStore() {
  const { subscribe, set } = writable<boolean>(false);

  return {
    subscribe,
    login: () => set(true),
    logout: () => set(false),
  };
}

export const isAuthenticated = createAuthStore();
export const username = writable<string>('');
