const fetchImages = () => new Promise((resolve) => {
  setTimeout(() => {
    resolve([
      {
        id: 'message-a7fe282a-bcb3-47a6-b746-e5f41fc2ea66',
        author: 'Robert',
        content: 'This is an awesome architecture',
        time: '2017-10-21T09:46:19+00:00',
        points: 5,
        reactions: [],
      },
    ]);
  }, 1000);
});


export default {
  fetchImages
}