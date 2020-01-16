from testdrive.resource import Status, Resource


class RunModel(object):
    def __init__(self, context):
        self.context = context
        self.services = {}
        self.services["workspace"] = Resource("container", "workspace",
                                              {"image": "gcr.io/google_containers/pause:3.0",
                                               "hostname": "workspace",
                                               "volumes": [{
                                                   "Source": context.workspace_path,
                                                   "Target": "/workspace",
                                                   "Type": "bind",
                                                   "ReadOnly": False
                                               }]},
                                              seqNum=context.next_seqnum(),
                                              dependencies=[])

    def set_driver(self, config):
        self.services["driver"] = Resource("container", "driver", config, seqNum=self.context.next_seqnum(),
                                           dependencies=["workspace"])

    def add_service(self, name, config):
        self.services[name] = Resource("container", name, config, seqNum=self.context.next_seqnum(),
                                       dependencies=["workspace"])

    def canStart(self, service):
        if service.status != Status.CREATED:
            return False

        for dependency in service.dependencies:
            if self.services[dependency].status not in [Status.READY]:
                return False

        return True

    def createServiceContainer(self, service):
        service.create(self.context)

    def startServiceContainer(self, service):
        service.start(self.context)

    def startHealthcheckForServiceContainer(self, service):
        service.startHealthcheck(self.context)

    def checkServiceContainer(self, service):
        service.check(self.context)

    def stopServiceContainer(self, service):
        service.stop(self.context)

    def killServiceContainer(self, service):
        service.kill(self.context)

    def removeServiceContainer(self, service):
        service.remove(self.context)
