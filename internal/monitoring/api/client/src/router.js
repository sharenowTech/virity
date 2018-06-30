import Vue from 'vue'
import Router from 'vue-router'
import Home from './views/Home.vue'
import About from './views/About.vue'
import CImage from './views/Image.vue'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/image/:id',
      name: 'cimage',
      component: CImage
    },
    {
      path: '/about',
      name: 'about',
      component: About
    }
  ],
  linkExactActiveClass: "is-active"
})
