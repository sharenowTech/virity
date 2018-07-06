//import api from '@/api'
import api from '@/api/local'

const defaultState = {
    images: [],
    list: []
};

const actions = {
    fetchImageList: (context) => {
        api.fetchImages()
            .then((response) => context.commit('LIST_UPDATED', response))
            .catch((error) => console.error(error))
    },
    fetchImageDetail: (context, params) => {
        api.fetchImageDetails(params.id)
            .then((response) => context.commit('IMAGE_UPDATED', response))
            .catch((error) => console.error(error))
    }
};

const mutations = {
    LIST_UPDATED: (state, images) => {
        state.list = images;
    },
    IMAGE_UPDATED: (state, image) => {
        // Change to Map if supported
        /*state.detail = {
            [image.id]: image,
            ...state.detail
        };*/
        state.images.push(image);
    },
};

const getters = {
    getImageById: (state) => (id) => {
        return state.images.find(image => image.id === id)
    }
};

export default {
    state: defaultState,
    getters,
    actions,
    mutations
};