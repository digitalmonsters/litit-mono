update notification set title = 'Real Video Uploaded'  where event_type = 'spot_upload';
update notification set body = 'Your real video was successfully uploaded' where event_type = 'spot_upload';