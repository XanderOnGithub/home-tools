import { mount } from 'svelte'
import '@home-tools/ui/app.css'
import { setAppTool } from '@home-tools/ui/tools'
import App from './App.svelte'

setAppTool('discord')

const app = mount(App, {
  target: document.getElementById('app')!,
})

export default app
