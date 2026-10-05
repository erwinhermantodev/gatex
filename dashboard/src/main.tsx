import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.tsx'
import './index.css'
import axios from 'axios'

// Admin API requires a bearer token; ask once and keep it in localStorage.
const TOKEN_KEY = 'gateway_admin_token'
axios.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})
axios.interceptors.response.use(undefined, (error) => {
  if (error.response?.status === 401) {
    const token = window.prompt('Admin API token')
    if (token) {
      localStorage.setItem(TOKEN_KEY, token)
      window.location.reload()
    }
  }
  return Promise.reject(error)
})

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
