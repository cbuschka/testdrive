const defaults = {
  services: {},
  drivers: {}
};

class Config {
  constructor(data) {
    this.data = Object.assign({}, defaults, data);
  }

  configure(dag) {
    Object.keys(this.data.services)
      .forEach((serviceId) => {
        const service = this.data.services[serviceId];
        service.name = serviceId;
        service.type = 'service';
        service.status = 'NEW';
        dag.addItem(serviceId, service);
      });
    Object.keys(this.data.drivers)
      .forEach((serviceId) => {
        const service = this.data.drivers[serviceId];
        service.name = serviceId;
        service.type = 'driver';
        service.status = 'NEW';
        dag.addItem(serviceId, service);
      });
  }
}

module.exports.Config = Config;

module.exports.example = {
  version: "testdrive:3.7",
  drivers: {
    hello: {
      run: {
        image: "debian:stretch-slim",
        command: ["echo", "hello"]
      },
    }
  },
  services: {
    db: {
      run: {
        image: "postgres:10"
      },
      check: {
        "docker-exec": {
          user: "postgres",
          command: "pgisready"
        }
      }
    },
  }
};
