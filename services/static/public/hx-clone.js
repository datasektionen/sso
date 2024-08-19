// Source: https://gist.github.com/pauloevpr/c0bae76e7e9e6e824953489184e899ce#file-hx-clone-js
htmx.config.useTemplateFragments = true
htmx.defineExtension('clone', {
    onEvent: function (name, evt) {
        if (name === 'htmx:beforeRequest') {
            if (evt.detail.elt) {
                const get = evt.detail.elt.getAttribute('hx-get')
                if (get && get.startsWith('clone-template#')) {
                    const selector = get.substring(15)
                    //console.log('htmx-clone: Intercepting xhr request to inject template with selector:', selector)
                    const template = document.querySelector(selector)
                    let templateContent = ''
                    if (!template) {
                        console.error(
                            'htmx-clone: No element found for selector: ' +
                                selector
                        )
                    } else {
                        if (template.tagName !== 'TEMPLATE') {
                            console.error(
                                'htmx-clone: Found element is not a <template>',
                                template
                            )
                        } else {
                            const templateNode =
                                template.content.cloneNode(true)
                            const tempDiv = document.createElement('div')
                            tempDiv.appendChild(templateNode)
                            templateContent = tempDiv.innerHTML
                        }
                    }
                    const xhr = evt.detail.xhr
                    Object.defineProperty(xhr, 'readyState', { writable: true })
                    Object.defineProperty(xhr, 'status', { writable: true })
                    Object.defineProperty(xhr, 'statusText', { writable: true })
                    Object.defineProperty(xhr, 'response', { writable: true })
                    Object.defineProperty(xhr, 'responseText', {
                        writable: true,
                    })
                    ;(xhr.readyState = 4),
                        (xhr.status = 200),
                        (xhr.statusText = 'OK'),
                        (xhr.response = templateContent),
                        (xhr.responseText = templateContent),
                        (xhr.send = () => {
                            xhr.onload()
                        })
                    xhr.clonedNode
                }
            }
        }
    },
})
