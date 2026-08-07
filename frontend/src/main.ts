import './style.css'
  import { mount } from 'svelte'
  import AppShell from './AppShell.svelte'

  const target = document.getElementById('app')

  if (!target) {
    throw new Error('missing #app element')
  }

  const app = mount(AppShell, { target })

  export default app
