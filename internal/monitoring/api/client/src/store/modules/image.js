//import api from '@/api'
import api from '@/api/local'

const defaultState = {
    images: [],
    detail: {},
};

const actions = {
    fetchImages: (context) => {
        api.fetchImages()
            .then((response) => context.commit('IMAGES_UPDATED', response))
            .catch((error) => console.error(error))
    },
    fetchDetails: (context, params) => {
        api.fetchImageDetails(params.id)
            .then((response) => context.commit('DETAILS_UPDATED', response))
            .catch((error) => console.error(error))
    }
};

const mutations = {
    IMAGES_UPDATED: (state, images) => {
        state.images = images;
    },
    DETAILS_UPDATED: (state, image) => {
        state.detail = {
            [image.id]: image,
            ...state.detail
        };
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