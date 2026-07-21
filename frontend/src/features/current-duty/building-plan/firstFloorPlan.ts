import type { FloorPlan } from './types'

const roomFill = '#F0FDFA'
const roomStroke = '#99F6E4'

export const firstFloorPlan: FloorPlan = {
    id: 'floor-1',
    floor: 1,
    name: '1 этаж',
    definition: {
        viewBox: {
            minX: 0,
            minY: 0,
            width: 1035,
            height: 542,
        },
        background: [
            {
                id: 'floor-1-shell',
                type: 'path',
                geometry: {
                    d: 'M0 0H1035V388H517.5H373V515.5H296.5V542H7V399H0V388V0Z',
                },
                fill: '#CBD5E1',
            },
            { id: 'floor-1-room-01', type: 'rect', geometry: { x: 21.5, y: 241.5, width: 142, height: 128, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-1-room-02', type: 'rect', geometry: { x: 693.5, y: 243.5, width: 152, height: 126, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-1-room-03', type: 'rect', geometry: { x: 851.5, y: 243.5, width: 159, height: 126, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-1-room-04', type: 'rect', geometry: { x: 28.5, y: 383.5, width: 135, height: 145, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-1-room-05', type: 'rect', geometry: { x: 112.5, y: 22.5, width: 107, height: 164, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-1-room-06', type: 'rect', geometry: { x: 225.5, y: 22.5, width: 111, height: 164, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            { id: 'floor-1-room-07', type: 'rect', geometry: { x: 693.5, y: 22.5, width: 128, height: 164, rx: 7.5 }, fill: roomFill, stroke: roomStroke },
            {
                id: 'floor-1-room-08',
                type: 'path',
                geometry: {
                    d: 'M842 22.5H911C915.142 22.5 918.5 25.8579 918.5 30V179C918.5 183.142 915.142 186.5 911 186.5H842C837.858 186.5 834.5 183.142 834.5 179V30C834.5 25.8579 837.858 22.5 842 22.5Z',
                },
                fill: roomFill,
                stroke: roomStroke,
            },
        ],
        areas: [
            {
                id: 'floor-1-area-01',
                areaId: '6',
                type: 'path',
                geometry: {
                    d: 'M48.75 187.5H74.75C77.6495 187.5 80 189.851 80 192.75C80 196.202 82.7982 199 86.25 199H125.25C128.702 199 131.5 196.202 131.5 192.75C131.5 189.851 133.851 187.5 136.75 187.5H154.75C157.65 187.5 160 189.851 160 192.75C160 196.202 162.798 199 166.25 199H286.75C290.202 199 293 196.202 293 192.75C293 189.851 295.351 187.5 298.25 187.5H313.75C316.649 187.5 319 189.851 319 192.75C319 196.202 321.798 199 325.25 199H360.5C364.642 199 368 202.358 368 206.5V229C368 233.142 364.642 236.5 360.5 236.5H298C293.306 236.5 289.5 240.306 289.5 245V362C289.5 366.142 286.142 369.5 282 369.5H232.25C227.97 369.5 224.5 372.97 224.5 377.25C224.5 380.978 221.478 384 217.75 384H178C173.858 384 170.5 380.642 170.5 376.5V248C170.5 243.306 166.694 239.5 162 239.5H29C24.8579 239.5 21.5 236.142 21.5 232V206.5C21.5 202.358 24.8579 199 29 199H37.25C40.7018 199 43.5 196.202 43.5 192.75C43.5 189.851 45.8505 187.5 48.75 187.5Z',
                },
                label: { x: 195, y: 220 },
            },
            { id: 'floor-1-area-02', areaId: '5', type: 'rect', geometry: { x: 369.5, y: 22.5, width: 145, height: 347, rx: 7.5 }, label: { x: 442, y: 196 } },
            { id: 'floor-1-area-03', areaId: '8', type: 'rect', geometry: { x: 515.5, y: 22.5, width: 145, height: 347, rx: 7.5 }, label: { x: 588, y: 196 } },
            {
                id: 'floor-1-area-04',
                areaId: '9',
                type: 'path',
                geometry: {
                    d: 'M714.75 188H734.25C736.597 188 738.5 189.903 738.5 192.25C738.5 195.149 740.851 197.5 743.75 197.5H878.75C881.649 197.5 884 195.149 884 192.25C884 189.903 885.903 188 888.25 188H906.75C909.097 188 911 189.903 911 192.25C911 195.149 913.351 197.5 916.25 197.5H948.25C951.149 197.5 953.5 195.149 953.5 192.25C953.5 189.976 955.286 188.12 957.531 188.006L957.75 188H984.75C987.097 188 989 189.903 989 192.25C989 195.149 991.351 197.5 994.25 197.5H1003C1007.14 197.5 1010.5 200.858 1010.5 205V230C1010.5 234.142 1007.14 237.5 1003 237.5H669C664.858 237.5 661.5 234.142 661.5 230V205C661.5 200.858 664.858 197.5 669 197.5H705.25C708.149 197.5 710.5 195.149 710.5 192.25C710.5 189.903 712.403 188 714.75 188Z',
                },
                label: { x: 836, y: 218 },
            },
            { id: 'floor-1-area-05', areaId: '', type: 'rect', geometry: { x: 295.5, y: 319.5, width: 43, height: 50, rx: 7.5 }, label: { x: 317, y: 346 } },
            {
                id: 'floor-1-area-06',
                areaId: '7',
                type: 'path',
                geometry: {
                    d: 'M303 244.5H331C335.142 244.5 338.5 247.858 338.5 252V307C338.5 311.142 335.142 314.5 331 314.5H303C298.858 314.5 295.5 311.142 295.5 307V252L295.51 251.614C295.711 247.651 298.987 244.5 303 244.5Z',
                },
                label: { x: 317, y: 280 },
            },
            { id: 'floor-1-area-07', areaId: '12', type: 'rect', geometry: { x: 21.5, y: 22.5, width: 80, height: 164, rx: 7.5 }, label: { x: 62, y: 106 } },
            { id: 'floor-1-area-08', areaId: 'floor-1-area-08', type: 'rect', geometry: { x: 930.5, y: 22.5, width: 80, height: 164, rx: 7.5 }, label: { x: 970, y: 106 } },
            {
                id: 'floor-1-area-09',
                areaId: '10',
                type: 'path',
                geometry: {
                    d: 'M178 384.5H267C271.142 384.5 274.5 387.858 274.5 392V429.5C274.5 434.194 278.306 438 283 438H287.5C292.194 438 296 434.194 296 429.5V396.5C296 392.358 299.358 389 303.5 389H357.5C361.642 389 365 392.358 365 396.5V497.5C365 501.642 361.642 505 357.5 505H283C278.306 505 274.5 508.806 274.5 513.5V520C274.5 524.142 271.142 527.5 267 527.5H178C173.858 527.5 170.5 524.142 170.5 520V392C170.5 387.858 173.858 384.5 178 384.5Z',
                },
                label: { x: 264, y: 458 },
            },
        ],
    },
}
