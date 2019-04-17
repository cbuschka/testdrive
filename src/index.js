const {Dag} = require('./dag');
const {example, Config} = require('./config');
const {DockerWatch} = require('./docker-watch');

const config = new Config(example);

const dag = new Dag();
config.configure(dag);

const {Docker} = require('node-docker-api');
const docker = new Docker({socketPath: '/var/run/docker.sock'});

const {Controller} = require('./controller');
const controller = new Controller(dag, docker);

const dockerWatch = new DockerWatch(docker, dag, (ev) => {
  controller.handle(ev);
});
dockerWatch.start();

controller.start();

  //docker.container.list()
  //  .then(containers => console.log(containers[0].data.Names[0]))
  /*
  .then(container => container.stats())
  .then(stats => {
    stats.on('data', stat => console.log('Stats: ', stat[name]))
    stats.on('error', err => console.log('Error: ', err))
  })
   */
  // .then(container => container.stats())
  /*
  .then(stats => {
    stats.on('data', stat => console.log('Stats: ', stat.toString()))
    stats.on('error', err => console.log('Error: ', err))
  })
   */
  // .catch(error => console.log(error));

//const config = require('./config.js');
//main(config);

