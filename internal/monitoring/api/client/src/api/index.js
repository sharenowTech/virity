//import Axios from 'axios';

/*const getImages = () => {
  return Axios.get('http://192.168.99.100:8080/api/')
    .then((response) => Promise.resolve(response.data))
    .catch((error) => Promise.reject(error));
}*/

const getImages = () => {
  return Promise.resolve("Das ist ein Test");
}

export default {
  getImages
}