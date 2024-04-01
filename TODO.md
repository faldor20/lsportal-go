# Missing features
- [ ] Recursive inclusions
  Currently we cannot have an inclusion within an inclusion within an exclusion eg: ` <div> ${`<li>`}<div>` 
# Performance improvments
- [ ] Currently we always send full text document updates, this could be improved.
  - We would have to benchmark how significant this would be. 
  - I imagine runing the inclusion detection is fairly slow, so we might be best off still only doing that once any incoming changes are applied and then making a new change that covers the entire region that had any changes applied
  - Maybe we mostly actually get single didCHange events, in which case we can just run the inclusion detection and then fetch the new change content 
  
