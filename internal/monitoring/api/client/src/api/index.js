import Axios from 'axios';

const fetchImages = () => {
  return Axios.get('http://127.0.0.1:8080/api/image/')
    .then((response) => Promise.resolve(response.data))
    .catch((error) => Promise.reject(error));
}

const fetchImageDetails = (id) => {
  return Axios.get('http://127.0.0.1:8080/api/image/'+id)
    .then((response) => Promise.resolve(response.data))
    .catch((error) => Promise.reject(error));
}


export default {
  fetchImages,
  fetchImageDetails
}