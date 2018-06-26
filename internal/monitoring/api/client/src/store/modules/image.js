//import api from '@/api'
import api from '@/api/local'

const defaultState = {
    images: "asd",
};

const actions = {
    fetchImages: (context) => {
        api.fetchImages()
            .then((response) => context.commit('IMAGES_UPDATED', response))
            .catch((error) => console.error(error))
    }
};

const mutations = {
    IMAGES_UPDATED: (state, images) => {
        state.images = images;
    },
};

const getters = {

};

export default {
    state: defaultState,
    getters,
    actions,
    mutations
};