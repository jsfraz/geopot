// tslint:disable
/**
 * geopot - extended StatsApi with dashboard endpoints
 */

import type { Observable } from 'rxjs';
import type { AjaxResponse } from 'rxjs/ajax';
import { BaseAPI } from '../runtime';
import type { OperationOpts } from '../runtime';
import type {
    ModelsConnection,
    ModelsHourlyStat,
    ModelsLatLng,
    ModelsNumberValue,
    ModelsStringsValue,
    ModelsTopEntry,
} from '../models';

/**
 * @export
 */
export class StatsApi extends BaseAPI {

    /**
     * Get all latitude and longitude pairs from the database
     */
    getAllLatLng(): Observable<Array<ModelsLatLng>>
    getAllLatLng(opts?: OperationOpts): Observable<AjaxResponse<Array<ModelsLatLng>>>
    getAllLatLng(opts?: OperationOpts): Observable<Array<ModelsLatLng> | AjaxResponse<Array<ModelsLatLng>>> {
        return this.request<Array<ModelsLatLng>>({
            url: '/api/stats/allLatLng',
            method: 'GET',
        }, opts?.responseOpts);
    };

    /**
     * Get count of unique countries
     */
    getUniqueCountryCount(): Observable<ModelsNumberValue>
    getUniqueCountryCount(opts?: OperationOpts): Observable<AjaxResponse<ModelsNumberValue>>
    getUniqueCountryCount(opts?: OperationOpts): Observable<ModelsNumberValue | AjaxResponse<ModelsNumberValue>> {
        return this.request<ModelsNumberValue>({
            url: '/api/stats/uniqueCountryCount',
            method: 'GET',
        }, opts?.responseOpts);
    };

    /**
     * Get count of unique IP addresses
     */
    getUniqueIPCount(): Observable<ModelsNumberValue>
    getUniqueIPCount(opts?: OperationOpts): Observable<AjaxResponse<ModelsNumberValue>>
    getUniqueIPCount(opts?: OperationOpts): Observable<ModelsNumberValue | AjaxResponse<ModelsNumberValue>> {
        return this.request<ModelsNumberValue>({
            url: '/api/stats/uniqueIpCount',
            method: 'GET',
        }, opts?.responseOpts);
    };

    /**
     * Get the number of connections in the last 24 hours
     */
    getLast24HourConnections(): Observable<ModelsNumberValue>
    getLast24HourConnections(opts?: OperationOpts): Observable<AjaxResponse<ModelsNumberValue>>
    getLast24HourConnections(opts?: OperationOpts): Observable<ModelsNumberValue | AjaxResponse<ModelsNumberValue>> {
        return this.request<ModelsNumberValue>({
            url: '/api/stats/last24HourConnections',
            method: 'GET',
        }, opts?.responseOpts);
    };

    /**
     * Get the server's own info
     */
    getServerInfo(): Observable<ModelsConnection>
    getServerInfo(opts?: OperationOpts): Observable<AjaxResponse<ModelsConnection>>
    getServerInfo(opts?: OperationOpts): Observable<ModelsConnection | AjaxResponse<ModelsConnection>> {
        return this.request<ModelsConnection>({
            url: '/api/stats/selfInfo',
            method: 'GET',
        }, opts?.responseOpts);
    };

    /**
     * Get total connection count
     */
    getTotalConnectionCount(): Observable<ModelsNumberValue>
    getTotalConnectionCount(opts?: OperationOpts): Observable<AjaxResponse<ModelsNumberValue>>
    getTotalConnectionCount(opts?: OperationOpts): Observable<ModelsNumberValue | AjaxResponse<ModelsNumberValue>> {
        return this.request<ModelsNumberValue>({
            url: '/api/stats/totalConnections',
            method: 'GET',
        }, opts?.responseOpts);
    };

    /**
     * Get hourly connection counts for the last N hours
     */
    getHourlyStats(hours?: number): Observable<Array<ModelsHourlyStat>>
    getHourlyStats(hours?: number, opts?: OperationOpts): Observable<AjaxResponse<Array<ModelsHourlyStat>>>
    getHourlyStats(hours: number = 24, opts?: OperationOpts): Observable<Array<ModelsHourlyStat> | AjaxResponse<Array<ModelsHourlyStat>>> {
        return this.request<Array<ModelsHourlyStat>>({
            url: `/api/stats/hourly?hours=${hours}`,
            method: 'GET',
        }, opts?.responseOpts);
    };

    /**
     * Get top N countries by connection count
     */
    getTopCountries(limit?: number): Observable<Array<ModelsTopEntry>>
    getTopCountries(limit?: number, opts?: OperationOpts): Observable<AjaxResponse<Array<ModelsTopEntry>>>
    getTopCountries(limit: number = 15, opts?: OperationOpts): Observable<Array<ModelsTopEntry> | AjaxResponse<Array<ModelsTopEntry>>> {
        return this.request<Array<ModelsTopEntry>>({
            url: `/api/stats/topCountries?limit=${limit}`,
            method: 'GET',
        }, opts?.responseOpts);
    };

    /**
     * Get top N usernames by usage count
     */
    getTopUsernames(limit?: number): Observable<Array<ModelsTopEntry>>
    getTopUsernames(limit?: number, opts?: OperationOpts): Observable<AjaxResponse<Array<ModelsTopEntry>>>
    getTopUsernames(limit: number = 10, opts?: OperationOpts): Observable<Array<ModelsTopEntry> | AjaxResponse<Array<ModelsTopEntry>>> {
        return this.request<Array<ModelsTopEntry>>({
            url: `/api/stats/topUsernames?limit=${limit}`,
            method: 'GET',
        }, opts?.responseOpts);
    };

    /**
     * Get top N passwords by usage count
     */
    getTopPasswords(limit?: number): Observable<Array<ModelsTopEntry>>
    getTopPasswords(limit?: number, opts?: OperationOpts): Observable<AjaxResponse<Array<ModelsTopEntry>>>
    getTopPasswords(limit: number = 10, opts?: OperationOpts): Observable<Array<ModelsTopEntry> | AjaxResponse<Array<ModelsTopEntry>>> {
        return this.request<Array<ModelsTopEntry>>({
            url: `/api/stats/topPasswords?limit=${limit}`,
            method: 'GET',
        }, opts?.responseOpts);
    };

    /**
     * Get the N most recent connection attempts
     */
    getRecentConnections(limit?: number): Observable<Array<ModelsConnection>>
    getRecentConnections(limit?: number, opts?: OperationOpts): Observable<AjaxResponse<Array<ModelsConnection>>>
    getRecentConnections(limit: number = 100, opts?: OperationOpts): Observable<Array<ModelsConnection> | AjaxResponse<Array<ModelsConnection>>> {
        return this.request<Array<ModelsConnection>>({
            url: `/api/stats/recentConnections?limit=${limit}`,
            method: 'GET',
        }, opts?.responseOpts);
    };
}
