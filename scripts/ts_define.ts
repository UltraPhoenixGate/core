const sdk:any = {};

sdk.on('data',({}:{
  sender: string,
  payload: any
}) => {
  console.log('data received');
})

sdk.send('data',{
  payload: {
    temp: 20,
    wind: 10
  }
})

sdk.on('event',({}:{
  sender: string,
  eventName: string,
  payload: any
}) => {
  console.log('event received');
})

sdk.emit('event',{
  eventName: 'findSomething',
  payload: {
    location: 'here'
  }
})

sdk.on('alert',({}:{
  sender: string,
  alert: string
}) => {
  console.log('alert received');
})

sdk.on('data',({payload,sender}:{
  sender: string,
  payload: any
}) => {
  if(payload.temp > 30){
    sdk.emit('alert',{
      alert: 'hot'
    })
  }
})