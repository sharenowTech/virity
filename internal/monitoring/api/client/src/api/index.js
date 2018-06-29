import Axios from 'axios';

const fetchImages = () => {
  return Axios.get('http://127.0.0.1/api/images/')
    .then((response) => Promise.resolve(response.data))
    .catch((error) => Promise.reject(error));
}

const fetchImageDetails = (id) => {
  return Axios.get('http://127.0.0.1/api/images/'+id)
    .then((response) => Promise.resolve(response.data))
    .catch((error) => Promise.reject(error));
}


export default {
  fetchImages,
  fetchImageDetails
}