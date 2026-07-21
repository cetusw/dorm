import type { FloorPlan } from './types'

const roomFill = '#F0FDFA'
const roomStroke = '#99F6E4'

export const thirdFloorPlan: FloorPlan = {
    id: 'floor-3',
    floor: 3,
    name: '3 этаж',
    definition: {
        viewBox: {
            minX: 0,
            minY: 0,
            width: 1035,
            height: 542,
        },
        background: [
            {
                id: 'floor-3-shell',
                type: 'path',
                geometry: {
                    d: 'M0 0H1035V388H517.5H296.5V542H7V399H0V388V0Z',
                },
                fill: '#CBD5E1',
            },
            { id: 'floor-3-room-01', type: 'rect', geometry: { x: 18.5, y: 241.5, width: 183, height: 128, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-02', type: 'rect', geometry: { x: 347.5, y: 241.5, width: 106, height: 128, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-03', type: 'rect', geometry: { x: 457.5, y: 241.5, width: 109, height: 128, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-04', type: 'rect', geometry: { x: 570.5, y: 241.5, width: 100, height: 128, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-05', type: 'rect', geometry: { x: 675.5, y: 241.5, width: 175, height: 128, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-06', type: 'rect', geometry: { x: 855.5, y: 241.5, width: 161, height: 128, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-07', type: 'rect', geometry: { x: 25.5, y: 382.5, width: 176, height: 147, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-08', type: 'rect', geometry: { x: 351.5, y: 19.5, width: 79, height: 164, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-09', type: 'rect', geometry: { x: 435.5, y: 19.5, width: 79, height: 164, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-10', type: 'rect', geometry: { x: 519.5, y: 19.5, width: 79, height: 164, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-11', type: 'rect', geometry: { x: 603.5, y: 19.5, width: 79, height: 164, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-12', type: 'rect', geometry: { x: 696.5, y: 19.5, width: 129, height: 164, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-13', type: 'rect', geometry: { x: 838.5, y: 19.5, width: 85, height: 164, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-14', type: 'rect', geometry: { x: 111.5, y: 19.5, width: 108, height: 164, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-3-room-15', type: 'rect', geometry: { x: 224.5, y: 19.5, width: 114, height: 164, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
        ],
        areas: [
            { id: 'floor-3-area-01', areaId: '12', type: 'rect', geometry: { x: 18.5, y: 19.5, width: 80, height: 164, rx: 7.5 }, label: { x: 58, y: 103 } },
            { id: 'floor-3-area-02', areaId: 'floor-3-area-02', type: 'rect', geometry: { x: 936.5, y: 19.5, width: 80, height: 164, rx: 7.5 }, label: { x: 976, y: 103 } },
            { id: 'floor-3-area-03', areaId: '2', type: 'rect', geometry: { x: 294.5, y: 302.5, width: 48, height: 67, rx: 7.5 }, label: { x: 318, y: 337 } },
            {
                id: 'floor-3-area-04',
                areaId: '1',
                type: 'path',
                geometry: {
                    d: 'M45.75 185H71.75C74.6495 185 77 187.351 77 190.25C77 193.702 79.7982 196.5 83.25 196.5H123.75C127.202 196.5 130 193.702 130 190.25C130 187.351 132.351 185 135.25 185H153.25C156.149 185 158.5 187.351 158.5 190.25C158.5 193.702 161.298 196.5 164.75 196.5H288.25C291.702 196.5 294.5 193.702 294.5 190.25C294.5 187.351 296.851 185 299.75 185H316.75C319.649 185 322 187.351 322 190.25C322 193.702 324.798 196.5 328.25 196.5H393.75C397.202 196.5 400 193.702 400 190.25C400 187.351 402.351 185 405.25 185H422.75C425.649 185 428 187.351 428 190.25V192C428 194.485 430.015 196.5 432.5 196.5C434.985 196.5 437 194.485 437 192V190.25C437 187.351 439.351 185 442.25 185H459.25C462.149 185 464.5 187.351 464.5 190.25C464.5 193.702 467.298 196.5 470.75 196.5H561.75C565.202 196.5 568 193.702 568 190.25C568 187.351 570.351 185 573.25 185H591.25C594.149 185 596.5 187.351 596.5 190.25V192.5C596.5 194.709 598.291 196.5 600.5 196.5C602.709 196.5 604.5 194.709 604.5 192.5V190.25C604.5 187.351 606.851 185 609.75 185H627.75C630.649 185 633 187.351 633 190.25C633 193.702 635.798 196.5 639.25 196.5H704.25C707.702 196.5 710.5 193.702 710.5 190.25C710.5 187.351 712.851 185 715.75 185H733.75C736.649 185 739 187.351 739 190.25C739 193.702 741.798 196.5 745.25 196.5H885.25C888.702 196.5 891.5 193.702 891.5 190.25C891.5 187.351 893.851 185 896.75 185H914.25C917.149 185 919.5 187.351 919.5 190.25C919.5 193.702 922.298 196.5 925.75 196.5H951.25C954.702 196.5 957.5 193.702 957.5 190.25C957.5 187.351 959.851 185 962.75 185H989.25C992.149 185 994.5 187.351 994.5 190.25C994.5 193.702 997.298 196.5 1000.75 196.5H1009.5C1013.64 196.5 1017 199.858 1017 204V229.5C1017 233.642 1013.64 237 1009.5 237H351.5C346.806 237 343 240.806 343 245.5V291.5C343 295.642 339.642 299 335.5 299H298.5C293.806 299 290 302.806 290 307.5V361.5C290 365.642 286.642 369 282.5 369H277.25C273.522 369 270.5 372.022 270.5 375.75V378.5C270.5 380.709 272.291 382.5 274.5 382.5C276.157 382.5 277.5 383.843 277.5 385.5V522C277.5 526.142 274.142 529.5 270 529.5H214C209.858 529.5 206.5 526.142 206.5 522V390C206.5 385.858 209.858 382.5 214 382.5H228.25C231.978 382.5 235 379.478 235 375.75C235 372.022 231.978 369 228.25 369H214C209.858 369 206.5 365.642 206.5 361.5V245.5C206.5 240.806 202.694 237 198 237H26C21.8579 237 18.5 233.642 18.5 229.5V204C18.5 199.858 21.8579 196.5 26 196.5H34.25C37.7018 196.5 40.5 193.702 40.5 190.25C40.5 187.351 42.8505 185 45.75 185Z',
                },
                label: { x: 515, y: 218 },
            },
        ],
    },
}
