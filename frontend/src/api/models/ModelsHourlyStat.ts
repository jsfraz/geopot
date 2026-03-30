// tslint:disable
/**
 * geopot
 */

/**
 * @export
 * @interface ModelsHourlyStat
 */
export interface ModelsHourlyStat {
    /**
     * ISO 8601 timestamp for the hour bucket
     * @type {string}
     */
    bucket: string;
    /**
     * @type {number}
     */
    count: number;
}
