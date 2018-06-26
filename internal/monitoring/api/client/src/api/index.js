import Axios from 'axios';

const fetchImages = () => {
  return Axios.get('http://127.0.0.1/api')
    .then((response) => Promise.resolve(response.data))
    .catch((error) => Promise.reject(error));
}


export default {
  fetchImages
}