const SUPPORTED_TAGS = new Set(['svg', 'g', 'rect', 'path', 'polygon'])
const FORBIDDEN_TAGS = ['script', 'foreignObject', 'image', 'animate', 'animateMotion', 'animateTransform', 'set']

function parseAttributes(tagSource: string): Record<string, string> {
    const attributes: Record<string, string> = {}
    const attributePattern = /([:@A-Za-z0-9_-]+)\s*=\s*"([^"]*)"/g

    for (const match of tagSource.matchAll(attributePattern)) {
        const [, name, value] = match
        if (name) {
            attributes[name] = value ?? ''
        }
    }

    return attributes
}

function fail(filePath: string, reason: string, elementId?: string): never {
    const idSuffix = elementId ? ` [element id="${elementId}"]` : ''
    throw new Error(`${filePath}${idSuffix}: ${reason}`)
}

function extractLayerMatch(svgSource: string, layerName: string): RegExpMatchArray | null {
    const pattern = new RegExp(
        `<g\\b([^>]*)data-plan-layer="${layerName}"([^>]*)>([\\s\\S]*?)<\\/g>`,
        'i',
    )

    return svgSource.match(pattern)
}

function validateViewBox(svgSource: string, filePath: string) {
    const svgTagMatch = svgSource.match(/<svg\b([^>]*)>/i)
    if (!svgTagMatch?.[1]) {
        fail(filePath, 'SVG root tag not found')
    }

    const attrs = parseAttributes(svgTagMatch[1])
    const viewBox = attrs.viewBox
    if (!viewBox) {
        fail(filePath, 'Missing viewBox attribute')
    }

    const parts = viewBox.split(/\s+/)
    if (parts.length !== 4 || parts.some((part) => Number.isNaN(Number(part)))) {
        fail(filePath, `Invalid viewBox value "${viewBox}"`)
    }
}

function validateNoForbiddenMarkup(svgSource: string, filePath: string) {
    for (const forbiddenTag of FORBIDDEN_TAGS) {
        const pattern = new RegExp(`<${forbiddenTag}\\b`, 'i')
        if (pattern.test(svgSource)) {
            fail(filePath, `Forbidden SVG element <${forbiddenTag}>`)
        }
    }

    if (/\b(?:href|xlink:href)\s*=/.test(svgSource)) {
        fail(filePath, 'External links are not allowed')
    }

    if (/\bon[a-z]+\s*=/.test(svgSource)) {
        fail(filePath, 'Inline event handlers are not allowed')
    }
}

function validateTagSet(svgSource: string, filePath: string) {
    const tagPattern = /<\/?([A-Za-z][A-Za-z0-9:-]*)\b/g

    for (const match of svgSource.matchAll(tagPattern)) {
        const tagName = match[1]
        if (!tagName) {
            continue
        }

        if (!SUPPORTED_TAGS.has(tagName)) {
            fail(filePath, `Unsupported SVG element <${tagName}>`)
        }
    }
}

function validateRequiredLayers(svgSource: string, filePath: string) {
    const backgroundLayer = extractLayerMatch(svgSource, 'background')
    const areasLayer = extractLayerMatch(svgSource, 'areas')

    if (!backgroundLayer) {
        fail(filePath, 'Missing required layer data-plan-layer="background"')
    }

    if (!areasLayer) {
        fail(filePath, 'Missing required layer data-plan-layer="areas"')
    }
}

function validateShapeAttributes(
    filePath: string,
    tagName: string,
    attrs: Record<string, string>,
    interactive: boolean,
) {
    const elementId = attrs.id

    if (!elementId) {
        fail(filePath, `Missing required id on <${tagName}>`)
    }

    if (interactive && !attrs['data-area-id']) {
        fail(filePath, `Missing required data-area-id on <${tagName}>`, elementId)
    }

    const numericAttrs =
        tagName === 'rect'
            ? ['x', 'y', 'width', 'height', 'rx']
            : []

    for (const numericAttr of numericAttrs) {
        const value = attrs[numericAttr]
        if (value === undefined || value === '') {
            if (numericAttr === 'rx') {
                continue
            }

            fail(filePath, `Missing numeric attribute "${numericAttr}"`, elementId)
        }

        if (Number.isNaN(Number(value))) {
            fail(filePath, `Invalid numeric attribute "${numericAttr}"="${value}"`, elementId)
        }
    }

    if (tagName === 'path' && !attrs.d) {
        fail(filePath, 'Missing path attribute "d"', elementId)
    }

    if (tagName === 'polygon' && !attrs.points) {
        fail(filePath, 'Missing polygon attribute "points"', elementId)
    }
}

function validateLayerShapes(layerSource: string, filePath: string, interactive: boolean) {
    const shapePattern = /<(rect|path|polygon)\b([^>]*)\/?>/gi
    const seenIds = new Set<string>()

    for (const match of layerSource.matchAll(shapePattern)) {
        const tagName = match[1]
        const rawAttributes = match[2]
        if (!tagName || rawAttributes === undefined) {
            continue
        }

        const attrs = parseAttributes(rawAttributes)
        validateShapeAttributes(filePath, tagName, attrs, interactive)

        if (seenIds.has(attrs.id)) {
            fail(filePath, 'Duplicate shape id', attrs.id)
        }

        seenIds.add(attrs.id)
    }
}

export function validateSvgPlan(svgSource: string, filePath: string) {
    validateNoForbiddenMarkup(svgSource, filePath)
    validateTagSet(svgSource, filePath)
    validateViewBox(svgSource, filePath)
    validateRequiredLayers(svgSource, filePath)

    const backgroundLayer = extractLayerMatch(svgSource, 'background')
    const areasLayer = extractLayerMatch(svgSource, 'areas')

    validateLayerShapes(backgroundLayer?.[3] ?? '', filePath, false)
    validateLayerShapes(areasLayer?.[3] ?? '', filePath, true)
}
