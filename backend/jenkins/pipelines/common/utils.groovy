#!/usr/bin/env groovy

def renderTemplate(templatefilepath, binding) {
  def text = readFile templatefilepath
  def engine = new groovy.text.GStringTemplateEngine()
  def template = engine.createTemplate(text).make(binding)
  return template.toString()
}

def prepareImage(String filename) {
  dir('backend/jenkins') {
    hasImage = sh (returnStatus: true,
      script: 'docker build --build-arg user=$(id -u -n) --build-arg uid=$(id -u) ' +
              '--build-arg gid=$(id -g) -f docker/' + filename + ' -t ' + filename + ':latest .')
    return (hasImage==0)
  }
}

return this
