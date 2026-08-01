import { DOMParser } from '@xmldom/xmldom'
import type { Document, Element } from '@xmldom/xmldom'

import type {
    AreaShape,
    BackgroundShape,
    FloorPlanGeometry,
    PathShape,
    RectShape,
    ViewBox,
} from '../../src/features/current-duty/building-plan/types.ts'

const FORBIDDEN_AREA_STYLE_ATTRIBUTES = [
    'fill',
    'stroke',
    'stroke-width',
    'style',
    'class',
    'opacity',
] as const

function fail(filePath: string, message: string): never {
    throw new Error(`${filePath}: ${message}`)
}

function parseXmlDocument(svgSource: string, filePath: string): Document {
    try {
        const document = new DOMParser().parseFromString(svgSource, 'image/svg+xml')

        if (!document.documentElement) {
            fail(filePath, 'invalid XML')
        }

        return document
    } catch {
        fail(filePath, 'invalid XML')
    }
}

function requireSvgRoot(document: Document, filePath: string): Element {
    const svgElement = document.documentElement

    if (!svgElement || svgElement.tagName !== 'svg') {
        fail(filePath, 'root element must be <svg>')
    }

    return svgElement
}

function getElementChildren(element: Element): Element[] {
    const children: Element[] = []

    for (let index = 0; index < element.childNodes.length; index += 1) {
        const childNode = element.childNodes.item(index)
        if (childNode?.nodeType === 1) {
            children.push(childNode as Element)
        }
    }

    return children
}

function requireAttribute(element: Element, attributeName: string, filePath: string): string {
    const value = element.getAttribute(attributeName)?.trim() ?? ''
    if (value === '') {
        fail(filePath, `<${element.tagName}> is missing "${attributeName}"`)
    }

    return value
}

function parseNumberAttribute(element: Element, attributeName: string, filePath: string): number {
    const value = requireAttribute(element, attributeName, filePath)
    const parsedValue = Number(value)

    if (!Number.isFinite(parsedValue)) {
        fail(filePath, `<${element.tagName}> has ${attributeName}="${value}"`)
    }

    return parsedValue
}

function parseOptionalNumberAttribute(
    element: Element,
    attributeName: string,
    filePath: string,
): number | undefined {
    const rawValue = element.getAttribute(attributeName)?.trim() ?? ''
    if (rawValue === '') {
        return undefined
    }

    const parsedValue = Number(rawValue)
    if (!Number.isFinite(parsedValue)) {
        fail(filePath, `<${element.tagName}> has ${attributeName}="${rawValue}"`)
    }

    return parsedValue
}

function parseViewBox(svgElement: Element, filePath: string): ViewBox {
    const rawViewBox = requireAttribute(svgElement, 'viewBox', filePath)
    const parts = rawViewBox.split(/\s+/)

    if (parts.length !== 4) {
        fail(filePath, `<svg> has invalid "viewBox"="${rawViewBox}"`)
    }

    const [minX, minY, width, height] = parts.map((part) => Number(part))
    if (![minX, minY, width, height].every(Number.isFinite)) {
        fail(filePath, `<svg> has invalid "viewBox"="${rawViewBox}"`)
    }

    if (width <= 0 || height <= 0) {
        fail(filePath, `<svg> has invalid "viewBox"="${rawViewBox}"`)
    }

    return { minX, minY, width, height }
}

function validateRootLayers(svgElement: Element, filePath: string): void {
    for (const childElement of getElementChildren(svgElement)) {
        if (childElement.tagName !== 'g') {
            fail(filePath, `unsupported root element <${childElement.tagName}>`)
        }

        const layerName = childElement.getAttribute('data-plan-layer')?.trim() ?? ''
        if (layerName === '') {
            fail(filePath, 'root <g> is missing "data-plan-layer"')
        }

        if (layerName !== 'background' && layerName !== 'areas') {
            fail(filePath, `unsupported root layer "${layerName}"`)
        }
    }
}

function findRequiredLayer(
    svgElement: Element,
    layerName: 'background' | 'areas',
    filePath: string,
): Element {
    const matchingLayers = getElementChildren(svgElement).filter(
        (childElement) => childElement.tagName === 'g' && childElement.getAttribute('data-plan-layer')?.trim() === layerName,
    )

    if (matchingLayers.length === 0) {
        fail(filePath, `missing root layer "${layerName}"`)
    }

    if (matchingLayers.length > 1) {
        fail(filePath, `duplicate root layer "${layerName}"`)
    }

    return matchingLayers[0]
}

function collectShapeElements(
    layerElement: Element,
    layerName: 'background' | 'areas',
    filePath: string,
): Element[] {
    const shapeElements: Element[] = []

    for (const childElement of getElementChildren(layerElement)) {
        if (childElement.tagName === 'g') {
            shapeElements.push(...collectShapeElements(childElement, layerName, filePath))
            continue
        }

        if (childElement.tagName !== 'rect' && childElement.tagName !== 'path') {
            fail(filePath, `unsupported <${childElement.tagName}> in "${layerName}" layer`)
        }

        shapeElements.push(childElement)
    }

    return shapeElements
}

function parseRect(element: Element, filePath: string): RectShape {
    const x = parseNumberAttribute(element, 'x', filePath)
    const y = parseNumberAttribute(element, 'y', filePath)
    const width = parseNumberAttribute(element, 'width', filePath)
    const height = parseNumberAttribute(element, 'height', filePath)
    const rx = parseOptionalNumberAttribute(element, 'rx', filePath)

    if (width <= 0) {
        fail(filePath, `<rect> has width="${String(width)}"`)
    }

    if (height <= 0) {
        fail(filePath, `<rect> has height="${String(height)}"`)
    }

    if (rx !== undefined && rx < 0) {
        fail(filePath, `<rect> has rx="${String(rx)}"`)
    }

    return {
        type: 'rect',
        x,
        y,
        width,
        height,
        rx,
    }
}

function parsePath(element: Element, filePath: string): PathShape {
    const d = requireAttribute(element, 'd', filePath)

    if (d.trim() === '') {
        fail(filePath, '<path> is missing "d"')
    }

    return {
        type: 'path',
        d,
    }
}

function parseShape(element: Element, filePath: string): RectShape | PathShape {
    switch (element.tagName) {
        case 'rect':
            return parseRect(element, filePath)
        case 'path':
            return parsePath(element, filePath)
        default:
            return fail(filePath, `unsupported <${element.tagName}>`)
    }
}

function parseBackgroundShape(element: Element, filePath: string): BackgroundShape {
    return {
        ...parseShape(element, filePath),
        fill: element.getAttribute('fill')?.trim() || undefined,
        stroke: element.getAttribute('stroke')?.trim() || undefined,
        strokeWidth: parseOptionalNumberAttribute(element, 'stroke-width', filePath),
    }
}

function parseAreaShape(element: Element, filePath: string): AreaShape {
    const areaId = requireAttribute(element, 'data-area-id', filePath)

    for (const attributeName of FORBIDDEN_AREA_STYLE_ATTRIBUTES) {
        if (element.hasAttribute(attributeName)) {
            fail(filePath, `area "${areaId}" must not define "${attributeName}"`)
        }
    }

    return {
        ...parseShape(element, filePath),
        areaId,
    }
}

export function parseSvgPlan(svgSource: string, filePath: string): FloorPlanGeometry {
    const document = parseXmlDocument(svgSource, filePath)
    const svgElement = requireSvgRoot(document, filePath)

    validateRootLayers(svgElement, filePath)

    const backgroundLayer = findRequiredLayer(svgElement, 'background', filePath)
    const areasLayer = findRequiredLayer(svgElement, 'areas', filePath)

    return {
        viewBox: parseViewBox(svgElement, filePath),
        background: collectShapeElements(backgroundLayer, 'background', filePath).map((element) =>
            parseBackgroundShape(element, filePath),
        ),
        areas: collectShapeElements(areasLayer, 'areas', filePath).map((element) =>
            parseAreaShape(element, filePath),
        ),
    }
}
