#!/bin/env python3

# stress test automation
#
# How to push images from directory
# python3 stress_test.py -action push -workload-id test -filter-name gray -token token -frames-path frames
#
# How to pull result images from API
# python3 stress_test.py -action pull -workload-id test  -token token -frames-path frames
#
# TODO
#
# - Add pull_images function
# - Add timing metrics
# - 


import argparse
import glob
import os
import requests

BASE_API_ENDPOINT='http://localhost:8080'
FILTER_ENDPOINT='workloads/filter'


# push_images sends images to the API to be filtered.
# images (frames) are taken from an specified directory
def push_images(frames_path, workload_id, filter_name, token):

    if not os.path.isdir(frames_path):
        print('[{}] frames path doesn\'t exist'.format(frames_path))
        return

    frames = glob.glob('{}/*.jpg'.format(frames_path))

    data = {'workload-id':workload_id, 'filter':filter_name}
    headers= {'Authorization': 'Bearer {}'.format(token)}
    endpoint_url = '{}/{}'.format(BASE_API_ENDPOINT, FILTER_ENDPOINT)

    for count in range(0,len(frames)):
        image_path = '{}/{}.jpg'.format(frames_path,count)
        files = {'data': open(image_path,'rb')}

        print('Sending {} frame'.format(image_path))
        r = requests.get(endpoint_url, files=files, headers=headers, data=data)
        print(endpoint_url)
        print(r.status_code)


# pull_images pulls results images
def pull_images(frames_path, workload_id, token):
    pass

if __name__ == '__main__':

    parser = argparse.ArgumentParser()
    parser.add_argument('-action', default='extract', help='extract or join video frames')
    parser.add_argument('-workload-id', default='test', help='Workload identifier')
    parser.add_argument('-filter-name', default='filter', help='Desired Filter')
    parser.add_argument('-token', default='token', help='API Token')
    parser.add_argument('-frames-path', default='frames', help='frames path')

    args = parser.parse_args()
    print(args)
    if args.action == 'push':
        push_images(args.frames_path, args.workload_id, args.filter_name, args.token)
    elif args.action == 'pull':
        pull_images(args.frames_path, args.workload_id, args.token)
