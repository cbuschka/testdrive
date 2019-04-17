class DockerWatch {
  constructor(docker, dag, handler) {
    this.docker = docker;
    this.dag = dag;
    this.handler = handler;
  }

  start() {
    this.docker
      .events({
        since: (new Date().getTime() / 1000).toFixed(0)
      })
      .then((stream) => {
          stream.on('data', (data) => {
            const event = JSON.parse(data.toString());

            switch (event.status) {
              case 'create':
                if (event.Actor && event.Actor.Attributes) {
                  this.handler({type: 'containerCreated', data: {name: event.Actor.Attributes.name}});
                }
                break;
              case 'start':
                if (event.Actor && event.Actor.Attributes) {
                  this.handler({type: 'containerStarted', data: {name: event.Actor.Attributes.name}});
                }
                break;
              case 'die':
                if (event.Actor && event.Actor.Attributes) {
                  this.handler({type: 'containerDied', data: {name: event.Actor.Attributes.name}});
                }
                break;
              default:
              // console.log("event %o", event);
            }
          });
          stream.on('end', () => {
            process.exit(0);
          });
          stream.on('error', () => {
            process.exit(1);
          });
        }
      )
      .catch((error) => console.log(error));
  }
}


module.exports.DockerWatch = DockerWatch;
