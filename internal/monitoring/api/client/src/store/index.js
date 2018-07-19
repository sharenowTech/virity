import Vue from 'vue'
import Vuex from 'vuex'
import imageModule from './modules/image'

Vue.use(Vuex)

export default new Vuex.Store({
  modules: {
    images: imageModule,
  },
});
