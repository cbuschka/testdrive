class Controller {

  constructor(dag, docker) {
    this.dag = dag;
    this.docker = docker;
    this.nextTick = this.nextTick.bind(this);
  }

  handle(ev) {
    console.log("event: %o", ev);
    switch (ev.type) {
      case 'containerCreated':
        break;
      case 'containerStarted':
        break;
      case 'containerDied':
        this.dag.getItem(ev.data.name).status = 'STOPPED';
        this.scheduleNextTick();
        break;
    }
  }

  start() {
    this.nextTick();
  }

  scheduleNextTick() {
    setImmediate(this.nextTick);
  }

  nextTick() {
    const actions = this.dag.getItems()
      .map((item) => {
        switch (item.status) {
          case 'NEW':
            return () => {
              this.createContainer(item);
            };
          case 'CREATED':
            return () => {
              this.startContainer(item);
            };
          case 'STARTED':
            return () => {
              this.pingContainer(item);
            };
          case 'RUNNING':
            return null;
          case 'STOPPED':
            return null;
          default:
            break;
        }

        return null;
      })
      .filter((item) => !!item);

    if (actions.length === 0) {
      const shutdownActions = this.dag.getItems()
        .filter((item) => item.status === 'RUNNING')
        .map((item) => {
          return () => {
            this.stopContainer(item);
          };
        })
        .filter((action) => !!action);
      if (shutdownActions.length === 0) {
        process.exit(0);
      } else {
        console.log("Shutting down...");
        shutdownActions.forEach((action) => action());
      }
    } else {
      actions.forEach((action) => {
        setImmediate(action)
      });
    }
  }

  createContainer(item) {
    item.status = 'CREATING';
    this.docker.container
      .create({
        Image: item.run.image,
        name: item.name,
        Cmd: item.run.command || []
      })
      .then((container) => {
        item.status = 'CREATED';
        item.container = container;
        this.scheduleNextTick();
      });
  }

  startContainer(item) {
    console.log("Starting container %o...", item.name);
    item.status = 'STARTING';
    item.container.start()
      .then((container) => {
        item.status = 'STARTED';
      });
  }

  stopContainer(item) {
    console.log("Stopping container %o...", item.name);
    item.status = 'STOPPING';
    item.container.kill()
      .then(() => {
        delete item.container;
      })
  }

  pingContainer(item) {
    item.status = 'RUNNING';
    this.scheduleNextTick();
  }
}

module.exports.Controller = Controller;
