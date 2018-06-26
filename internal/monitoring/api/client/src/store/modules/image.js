import api from '@/api'

const defaultState = {
    images: "asd",
};

const actions = {
    getImages: (context) => {
        api.getImages()
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