class Dag {
  constructor() {
    this.nodes = {};
  }

  addItem(id, item, dependencies) {
    this.nodes[id] = {id, item, dependencies: dependencies || {}};
  }

  getItem(id) {
    return this.nodes[id].item;
  }

  getItems() {
    return Object.keys(this.nodes)
      .map((id) => this.nodes[id].item);
  }

  getDependencyItemsOf(id) {
    return Object.keys(this.nodes[i].dependencies)
      .map((id) => this.nodes[i] ? this.nodes[id].item : null)
      .filter((item) => !!item);
  }
}

module.exports.Dag = Dag;
